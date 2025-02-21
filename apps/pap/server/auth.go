// Package server contains the HTTP request handlers for the app.
package server

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pap/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pap/persistence"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/cedar"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/cerbos"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/opa"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/openfga"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
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
	l := models.LanguageFromString(cfg.PolicyLanguage)

	pipCfg := pip.Config{Ctx: ctx, Store: cfg.PipStore, Recurse: cfg.PipStoreRecurse, Logger: logger, PullConfigs: cfg.PipPullConfigs}
	if l == models.CEDAR {
		pipCfg.NewAttributes, pipCfg.NewEntities = cedar.NewAttributeBuilder(logger), cedar.NewEntityBuilder(logger)
	}
	p1 := pip.New(pipCfg)

	papOpts := []pap.Option{pap.WithLanguage(l.Language()), pap.WithFileStore(cfg.PolicyStore, cfg.PolicyStoreRecurse)}
	if cfg.PersistType != "" {
		s, err := persistence.New(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create persistence store: %w", err)
		}
		papOpts = append(papOpts, pap.WithPersistence(s, cfg.PersistBase))
	}
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

type authHandler struct {
	logger     *slog.Logger
	controller pdp.Controller
}
