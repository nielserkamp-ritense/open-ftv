package server

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"

	handlers "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cerbos-api"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	odrl_geo "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl-geo"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/opa-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/openfga-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	_ "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/plugins/mapping-body"
	_ "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/plugins/mapping-doelbinding"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/opentelemetry"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config"
)

// AuthHandler represents the interface for handling authorization requests.
type AuthHandler interface {
	Controller() pdp.Controller
	AuthZEN(req *fiber.Ctx) error
}

func (s *service) newAuth() AuthHandler {
	ip, err := s.cfg.PIP.NewPIP(s.ctx, s.logger, s.l)
	if err != nil {
		s.logger.Error("failed to initialize pip", "error", err)
		return nil
	}

	controller, err := s.newController(ip)
	if controller == nil {
		s.logger.Error("failed to initialize pdp controller", "error", err)
		return nil
	}

	authLogger, err := s.newAuthLogger(ip)
	if err != nil {
		s.logger.Error("failed to initialize authlog", "error", err)
		return nil
	}

	zen := handlers.NewAuthHandlerZEN(s.logger, authLogger, controller)

	return &authHandler{logger: s.logger, controller: controller, zen: zen}
}

// newAuthLogger selects the decision-log implementation: the ADL-conformant logger when
// enabled (with a durable WAL and asynchronous OTLP and/or OpenSearch sinks), otherwise
// the legacy OpenSearch logger, otherwise none.
func (s *service) newAuthLogger(ip pip.PIP) (authlog.Logger, error) {
	if !s.cfg.ADL.Enabled {
		if s.cfg.OpenSearch.Index == "" {
			return nil, nil
		}
		return authlog.NewOpenSearch(s.cfg.OpenSearch.Index, s.cfg.OpenSearch.User, s.cfg.OpenSearch.Pswd, strings.Split(s.cfg.OpenSearch.Endpoints, ",")...)
	}

	// Gate the ADL TLS ("Encryption") requirement before constructing any sink: the
	// sink constructors reject cleartext http:// to non-loopback unless this is set.
	adl.AllowInsecureTransport = s.cfg.ADL.Insecure

	var sinks []adl.Sink

	if s.cfg.ADL.OTLPURL != "" {
		sink, err := adl.NewOTLPSink(&opentelemetry.LoggerConfig{Service: adlServiceName, URL: s.cfg.ADL.OTLPURL, Logger: s.logger})
		if err != nil {
			return nil, err
		}
		sinks = append(sinks, sink)
	}

	if s.cfg.OpenSearch.Index != "" {
		sink, err := adl.NewOpenSearchSink(s.cfg.OpenSearch.Index, s.cfg.OpenSearch.User, s.cfg.OpenSearch.Pswd, strings.Split(s.cfg.OpenSearch.Endpoints, ",")...)
		if err != nil {
			return nil, err
		}
		sinks = append(sinks, sink)
	}

	logger, err := adl.New(adl.Config{
		Level:           s.cfg.ADL.Level,
		WALPath:         s.cfg.ADL.Path,
		Resource:        s.cfg.ADL.ResourceMap(map[string]any{"service.name": adlServiceName}),
		Engine:          s.cfg.PAP.Language,
		EngineVersion:   config.AppName,
		ConfigHash:      s.cfg.ADL.ConfigHash,
		RequestMappings: s.cfg.RequestMappings,
		PIPStore:        s.cfg.PIP.Store,
		PIPPull:         s.cfg.PIP.PullConfigs,
		PIPWARC:         s.cfg.PIP.WARCDir,
		PAPStore:        s.cfg.PAP.Store,
		Information:     newInformationProvider(ip),
		Logger:          s.logger,
	}, sinks...)
	if err != nil {
		return nil, err
	}

	// Ingreep 5: WAL flush-recovery. Re-emit any records the previous run appended to
	// the durable WAL but may not have flushed to the sinks. Emit is idempotent on
	// (trace_id, span_id), so replaying already-flushed records produces no duplicates.
	if err := logger.Replay(s.ctx); err != nil {
		s.logger.Warn("adl: write-ahead log replay failed", "error", err)
	}

	return logger, nil
}

// adlServiceName identifies this PDP as the producer of ADL records.
const adlServiceName = "pdp"

func (s *service) newController(ip pip.PIP) (pdp.Controller, error) {
	ep := pep.New(s.ctx, s.logger)

	ap, err := s.cfg.PAP.NewPAP(s.ctx, s.logger)
	if err != nil {
		return nil, err
	}

	options := []pdp.Option{pdp.WithContext(s.ctx), pdp.WithLogger(s.logger), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap)}
	if s.cfg.RequestMappings != "" {
		mappers, unknown := mapping.Resolve(s.cfg.RequestMappings)
		if len(unknown) > 0 {
			s.logger.Warn("request mapping(s) geconfigureerd maar niet geladen — plugin niet geïmporteerd?", "unknown", unknown)
		}
		options = append(options, pdp.WithMappings(mappers...))
	}

	switch s.l {
	case models.ODRL, models.ODRLGEO:
		// The ODRL-Geo-NL engine consumes policies stored under the "odrl"
		// PAP language, so PAP_LANGUAGE=odrl and =odrl-geo both select it.
		return odrl_geo.NewController(options...), nil
	case models.CEDAR:
		return cedar_embedded.NewController(options...), nil
	case models.REGO:
		return opa_embedded.NewController(options...), nil
	case models.OPENFGA:
		return openfga_embedded.NewController(options...), nil
	case models.CERBOS:
		cerbosCFG := cerbos_api.Config{Addr1: s.cfg.Cerbos.Address, Addr2: s.cfg.Cerbos.AdminAddress, CA: s.cfg.Cerbos.CA, User: s.cfg.Cerbos.User, Pswd: s.cfg.Cerbos.Pswd}
		return cerbos_api.NewController(cerbosCFG, options...), nil
	default:
		return nil, fmt.Errorf("unsupported policy language '%s'", s.cfg.PAP.Language)
	}
}

// Controller returns the PDP controller.
func (h *authHandler) Controller() pdp.Controller { return h.controller }

// AuthZEN authorizes an AuthZEN authorization request.
func (h *authHandler) AuthZEN(req *fiber.Ctx) error { return h.zen.Authorize(req) }

type authHandler struct {
	logger     *slog.Logger
	controller pdp.Controller
	zen        handlers.AuthZENAuthorizer
}
