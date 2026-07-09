// Package server contains the HTTP request handlers for the FSC Auth plugin.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"

	handlers "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cerbos-api"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/opa-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/openfga-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	_ "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/plugins/mapping-body"
	_ "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/plugins/mapping-doelbinding"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/opentelemetry"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/fsc-auth/config"
)

// AuthHandler represents the interface for handling authorization requests.
type AuthHandler interface {
	Controller() pdp.Controller
	AuthFSC(req *fiber.Ctx) error
	AuthZEN(req *fiber.Ctx) error
}

// New instantiates an authorization handler.
func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) AuthHandler {
	if ctx == nil {
		ctx = context.Background()
	}

	l := models.LanguageFromString(cfg.PAP.Language)

	ip, err := cfg.PIP.NewPIP(ctx, logger, l)
	if err != nil {
		logger.Error("failed to initialize pip", "error", err)
		return nil
	}

	controller, err := newController(ctx, cfg, logger, l, ip)
	if controller == nil {
		logger.Error("failed to initialize pdp controller", "error", err)
		return nil
	}

	authLogger, err := newAuthLogger(ctx, cfg, logger, ip)
	if err != nil {
		logger.Error("failed to initialize authlog", "error", err)
		return nil
	}

	fsc := handlers.NewAuthHandlerFSC(logger, authLogger, controller)
	zen := handlers.NewAuthHandlerZEN(logger, authLogger, controller)

	return &authHandler{logger: logger, controller: controller, fsc: fsc, zen: zen}
}

// adlServiceName identifies this PDP as the producer of ADL records.
const adlServiceName = "fsc-auth"

// newAuthLogger selects the decision-log implementation: the ADL-conformant logger when
// enabled (with a durable WAL and asynchronous OTLP and/or OpenSearch sinks), otherwise
// the legacy OpenSearch logger, otherwise none.
func newAuthLogger(ctx context.Context, cfg *config.Config, logger *slog.Logger, ip pip2.PIP) (authlog.Logger, error) {
	if !cfg.ADL.Enabled {
		if cfg.OpenSearch.Index == "" {
			return nil, nil
		}
		return authlog.NewOpenSearch(cfg.OpenSearch.Index, cfg.OpenSearch.User, cfg.OpenSearch.Pswd, strings.Split(cfg.OpenSearch.Endpoints, ",")...)
	}

	// Gate the ADL TLS ("Encryption") requirement before constructing any sink: the
	// sink constructors reject cleartext http:// to non-loopback unless this is set.
	adl.AllowInsecureTransport = cfg.ADL.Insecure

	var sinks []adl.Sink

	if cfg.ADL.OTLPURL != "" {
		sink, err := adl.NewOTLPSink(&opentelemetry.LoggerConfig{Service: adlServiceName, URL: cfg.ADL.OTLPURL, Logger: logger})
		if err != nil {
			return nil, err
		}
		sinks = append(sinks, sink)
	}

	if cfg.OpenSearch.Index != "" {
		sink, err := adl.NewOpenSearchSink(cfg.OpenSearch.Index, cfg.OpenSearch.User, cfg.OpenSearch.Pswd, strings.Split(cfg.OpenSearch.Endpoints, ",")...)
		if err != nil {
			return nil, err
		}
		sinks = append(sinks, sink)
	}

	adlLogger, err := adl.New(adl.Config{
		Level:           cfg.ADL.Level,
		WALPath:         cfg.ADL.Path,
		Resource:        cfg.ADL.ResourceMap(map[string]any{"service.name": adlServiceName}),
		Engine:          cfg.PAP.Language,
		EngineVersion:   config.AppName,
		ConfigHash:      cfg.ADL.ConfigHash,
		RequestMappings: cfg.RequestMappings,
		PIPStore:        cfg.PIP.Store,
		PIPPull:         cfg.PIP.PullConfigs,
		PIPWARC:         cfg.PIP.WARCDir,
		PAPStore:        cfg.PAP.Store,
		Information:     newInformationProvider(ip),
		Logger:          logger,
	}, sinks...)
	if err != nil {
		return nil, err
	}

	// Ingreep 5: WAL flush-recovery on startup (idempotent on trace_id/span_id).
	if err := adlLogger.Replay(ctx); err != nil {
		logger.Warn("adl: write-ahead log replay failed", "error", err)
	}

	return adlLogger, nil
}

func newController(ctx context.Context, cfg *config.Config, logger *slog.Logger, l models.Language, ip pip2.PIP) (pdp.Controller, error) {
	ep := pep.New(ctx, logger)

	papOpts := []pap2.Option{pap2.WithLanguage(l.Language()), pap2.WithFileStore(cfg.PAP.Store, cfg.PAP.StoreRecurse)}
	ap := pap2.New(ctx, logger, papOpts...)

	pdpOpts := []pdp.Option{pdp.WithContext(ctx), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger)}
	if cfg.RequestMappings != "" {
		mappers, unknown := mapping.Resolve(cfg.RequestMappings)
		if len(unknown) > 0 {
			logger.Warn("request mapping(s) geconfigureerd maar niet geladen — plugin niet geïmporteerd?", "unknown", unknown)
		}
		pdpOpts = append(pdpOpts, pdp.WithMappings(mappers...))
	}

	switch l {
	case models.CEDAR:
		return cedar_embedded.NewController(pdpOpts...), nil
	case models.REGO:
		return opa_embedded.NewController(pdpOpts...), nil
	case models.OPENFGA:
		return openfga_embedded.NewController(pdpOpts...), nil
	case models.CERBOS:
		cerbosCFG := cerbos_api.Config{Addr1: cfg.Cerbos.Address, Addr2: cfg.Cerbos.AdminAddress, CA: cfg.Cerbos.CA, User: cfg.Cerbos.User, Pswd: cfg.Cerbos.Pswd}
		return cerbos_api.NewController(cerbosCFG, pdpOpts...), nil
	default:
		return nil, fmt.Errorf("unsupported policy language '%s'", cfg.PAP.Language)
	}
}

// Controller returns the PDP controller.
func (h *authHandler) Controller() pdp.Controller { return h.controller }

// AuthFSC handles an FSC Authorization request.
func (h *authHandler) AuthFSC(req *fiber.Ctx) error { return h.fsc.Authorize(req) }

// AuthZEN authorizes an AuthZEN authorization request.
func (h *authHandler) AuthZEN(req *fiber.Ctx) error { return h.zen.Authorize(req) }

type authHandler struct {
	logger     *slog.Logger
	controller pdp.Controller
	fsc        handlers.FSCAuthorizer
	zen        handlers.AuthZENAuthorizer
}
