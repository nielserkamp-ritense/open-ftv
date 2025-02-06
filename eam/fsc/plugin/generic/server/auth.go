// Package handlers contains the HTTP request handlers for the FSC Auth plugin.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/ldv"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/cedar"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/cerbos"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/opa"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/openfga"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	handlers "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers/fiber"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/fsc/plugin/generic/config"
)

// AuthHandler represents the interface for handling authorization requests.
type AuthHandler interface {
	Controller() pdp.Controller
	AuthFSC(req *fiber.Ctx) error
	AuthZEN(req *fiber.Ctx) error
}

// New instantiates an authorization handler.
func New(ctx context.Context, cfg *config.Config, logger *slog.Logger, logboek ldv.LDV) AuthHandler {
	controller, err := newController(ctx, cfg, logger, logboek)
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

func newController(ctx context.Context, cfg *config.Config, logger *slog.Logger, logboek ldv.LDV) (pdp.Controller, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	l := components.LanguageFromString(cfg.PolicyLanguage)

	var p pip.PIP
	switch l {
	case components.CEDAR:
		p = pip.New(pip.Config{Ctx: ctx, Store: cfg.PipStore, Recurse: cfg.PipStoreRecurse, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
	case components.REGO:
		p = pip.New(pip.Config{Ctx: ctx, Store: cfg.PipStore, Recurse: cfg.PipStoreRecurse, Logger: logger})
	case components.OPENFGA:
		p = pip.New(pip.Config{Ctx: ctx, Store: cfg.PipStore, Recurse: cfg.PipStoreRecurse, Logger: logger})
	case components.CERBOS:
		p = pip.New(pip.Config{Ctx: ctx, Store: cfg.PipStore, Recurse: cfg.PipStoreRecurse, Logger: logger})
	default:
	}

	options := []pdp.Option{pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore(cfg.PolicyStore, cfg.PolicyStoreRecurse), pdp.WithLogger(logger), pdp.WithLogboek(logboek)}

	switch l {
	case components.CEDAR:
		return cedar.NewController(options...), nil
	case components.REGO:
		return opa.NewController(options...), nil
	case components.OPENFGA:
		return openfga.NewController(options...), nil
	case components.CERBOS:
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
