// Package server contains the HTTP request handlers for the app.
package server

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pip/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/cedar"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/cerbos"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/opa"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/openfga"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
)

// AuthHandler represents the interface for handling authorization requests.
type AuthHandler interface {
	Controller() pdp.Controller
}

// New instantiates an authorization handler.
func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) AuthHandler {
	controller, err := newController(ctx, cfg, logger)
	if controller == nil {
		logger.Error("failed to initialize EAM controller", "error", err)
		return nil
	}

	return &authHandler{logger: logger, controller: controller}
}

func newController(ctx context.Context, cfg *config.Config, logger *slog.Logger) (pdp.Controller, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	l := components.LanguageFromString(cfg.PolicyLanguage)

	pipCfg := pip.Config{Ctx: ctx, Store: cfg.PipStore, Recurse: cfg.PipStoreRecurse, Logger: logger, PullConfigs: cfg.PipPullConfigs}

	var p pip.PIP
	switch l {
	case components.CEDAR:
		pipCfg.NewAttributes, pipCfg.NewEntities = cedar.NewAttributeBuilder(logger), cedar.NewEntityBuilder(logger)
		p = pip.New(pipCfg)
	case components.REGO:
		p = pip.New(pipCfg)
	case components.OPENFGA:
		p = pip.New(pipCfg)
	case components.CERBOS:
		p = pip.New(pipCfg)
	default:
	}

	options := []pdp.Option{pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore(cfg.PolicyStore, cfg.PolicyStoreRecurse), pdp.WithLogger(logger)}

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

type authHandler struct {
	logger     *slog.Logger
	controller pdp.Controller
}
