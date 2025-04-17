// Package server contains the HTTP request handlers for the FSC Auth plugin.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/cedar"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/cerbos"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/mapping"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/opa"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/openfga"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	handlers "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/fsc-auth/config"
)

// AuthHandler represents the interface for handling authorization requests.
type AuthHandler interface {
	Controller() pdp.Controller
	AuthFSC(req *fiber.Ctx) error
	AuthZEN(req *fiber.Ctx) error
}

// New instantiates an authorization handler.
func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) AuthHandler {
	controller, err := newController(ctx, cfg, logger)
	if controller == nil {
		logger.Error("failed to initialize pdp controller", "error", err)
		return nil
	}

	var authLogger authlog.Logger
	if cfg.OpenSearch.Index != "" {
		authLogger, err = authlog.NewOpenSearch(cfg.OpenSearch.Index, cfg.OpenSearch.User, cfg.OpenSearch.Pswd, strings.Split(cfg.OpenSearch.Endpoints, ",")...)
		if err != nil {
			logger.Error("failed to initialize authlog", "index", cfg.OpenSearch.Index, "user", cfg.OpenSearch.User, "endpoints", cfg.OpenSearch.Endpoints, "error", err)
			return nil
		}
	}

	fsc := handlers.NewAuthHandlerFSC(logger, authLogger, controller)
	zen := handlers.NewAuthHandlerZEN(logger, authLogger, controller)

	return &authHandler{logger: logger, controller: controller, fsc: fsc, zen: zen}
}

func newController(ctx context.Context, cfg *config.Config, logger *slog.Logger) (pdp.Controller, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	l := models.LanguageFromString(cfg.PAP.Language)

	ep := pep.New(ctx, logger)

	pipOpts := []pip.Option{pip.WithFileStore(cfg.PIP.Store, cfg.PIP.StoreRecurse), pip.WithPullConfigs(cfg.PIP.PullConfigs)}
	if l == models.CEDAR {
		pipOpts = append(pipOpts, pip.WithFactories(cedar.NewAttributeBuilder(logger), cedar.NewEntityBuilder(logger)))
	}
	ip := pip.New(ctx, logger, pipOpts...)

	papOpts := []pap.Option{pap.WithLanguage(l.Language()), pap.WithFileStore(cfg.PAP.Store, cfg.PAP.StoreRecurse)}
	ap := pap.New(ctx, logger, papOpts...)

	pdpOpts := []pdp.Option{pdp.WithContext(ctx), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger)}
	if cfg.RequestMappings != "" {
		pdpOpts = append(pdpOpts, pdp.WithMappings(mapping.MappingsFromConfig(cfg.RequestMappings)...))
	}

	switch l {
	case models.CEDAR:
		return cedar.NewController(pdpOpts...), nil
	case models.REGO:
		return opa.NewController(pdpOpts...), nil
	case models.OPENFGA:
		return openfga.NewController(pdpOpts...), nil
	case models.CERBOS:
		cerbosCFG := cerbos.Config{Addr1: cfg.Cerbos.Address, Addr2: cfg.Cerbos.AdminAddress, CA: cfg.Cerbos.CA, User: cfg.Cerbos.User, Pswd: cfg.Cerbos.Pswd}
		return cerbos.NewController(cerbosCFG, pdpOpts...), nil
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
