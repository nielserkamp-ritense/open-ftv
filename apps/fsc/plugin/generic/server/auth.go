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
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/opa"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/openfga"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	handlers "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/fsc/plugin/generic/config"
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
		logger.Error("failed to initialize EAM controller", "error", err)
		return nil
	}

	var authLogger authlog.Logger
	if cfg.OpenSearchIndex != "" {
		authLogger, err = authlog.NewOpenSearch(cfg.OpenSearchIndex, cfg.OpenSearchUser, cfg.OpenSearchPswd, strings.Split(cfg.OpenSearchEndpoints, ",")...)
		if err != nil {
			logger.Error("failed to initialize authlog", "index", cfg.OpenSearchIndex, "user", cfg.OpenSearchUser, "endpoints", cfg.OpenSearchEndpoints, "error", err)
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

	l := models.LanguageFromString(cfg.PolicyLanguage)

	pipCfg := pip.Config{Ctx: ctx, Store: cfg.PipStore, Recurse: cfg.PipStoreRecurse, Logger: logger, PullConfigs: cfg.PipPullConfigs}
	if l == models.CEDAR {
		pipCfg.NewAttributes, pipCfg.NewEntities = cedar.NewAttributeBuilder(logger), cedar.NewEntityBuilder(logger)
	}
	p1 := pip.New(pipCfg)

	papOpts := []pap.Option{pap.WithLanguage(l.Language()), pap.WithFileStore(cfg.PolicyStore, cfg.PolicyStoreRecurse)}
	p2 := pap.New(ctx, logger, papOpts...)

	options := []pdp.Option{pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger)}

	switch l {
	case models.CEDAR:
		return cedar.NewController(options...), nil
	case models.REGO:
		return opa.NewController(options...), nil
	case models.OPENFGA:
		return openfga.NewController(options...), nil
	case models.CERBOS:
		cerbosCFG := cerbos.Config{Addr1: cfg.CerbosAddress, Addr2: cfg.CerbosAdmin, CA: cfg.CerbosCA, User: cfg.CerbosUser, Pswd: cfg.CerbosPswd}
		return cerbos.NewController(cerbosCFG, options...), nil
	default:
		return nil, fmt.Errorf("unsupported policy language '%s'", cfg.PolicyLanguage)
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
