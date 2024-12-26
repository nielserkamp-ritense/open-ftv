// Package handlers contains the HTTP request handlers for the FSC Auth plugin.
package handlers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/ldv"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp/cedar"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp/cerbos"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp/opa"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp/openfga"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/config"
)

// AuthHandler represents the interface for handling authorization requests.
type AuthHandler interface {
	Controller() pdp.Controller
	AuthFSC(req *fiber.Ctx) error
	AuthZEN(req *fiber.Ctx) error
}

// New instantiates an authorization handler.
func New(ctx context.Context, cfg *config.Config, logger *slog.Logger, logboek ldv.LDV) AuthHandler {
	h, err := newController(ctx, cfg, logger, logboek)
	if h == nil {
		logger.Error("configuration error", "error", err)
		return nil
	}

	return &authHandler{cfg: cfg, logger: logger, controller: h}
}

func (h *authHandler) Controller() pdp.Controller {
	return h.controller
}

func newController(ctx context.Context, cfg *config.Config, logger *slog.Logger, logboek ldv.LDV) (pdp.Controller, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var p pip.PIP

	l := components.LanguageFromString(cfg.PolicyLanguage)
	switch l {
	case components.REGO:
		p = pip.New(pip.Config{
			Ctx:     ctx,
			Store:   cfg.PipStore,
			Recurse: cfg.PipStoreRecurse,
			Logger:  logger,
		})
	case components.CERBOS:
		p = pip.New(pip.Config{
			Ctx:     ctx,
			Store:   cfg.PipStore,
			Recurse: cfg.PipStoreRecurse,
			Logger:  logger,
		})
	case components.CEDAR:
		p = pip.New(pip.Config{
			Ctx:           ctx,
			Store:         cfg.PipStore,
			Recurse:       cfg.PipStoreRecurse,
			Logger:        logger,
			NewAttributes: cedar.NewAttributeBuilder(logger),
			NewEntities:   cedar.NewEntityBuilder(logger),
		})
	case components.OPENFGA:
		p = pip.New(pip.Config{
			Ctx:     ctx,
			Store:   cfg.PipStore,
			Recurse: cfg.PipStoreRecurse,
			Logger:  logger,
		})
	default:
	}

	options := []pdp.Option{pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore(cfg.PolicyStore, cfg.PolicyStoreRecurse), pdp.WithLogger(logger), pdp.WithLogboek(logboek)}

	switch l {
	case components.REGO:
		return opa.NewController(options...), nil
	case components.CERBOS:
		return cerbos.NewController(options...), nil
	case components.CEDAR:
		return cedar.NewController(options...), nil
	case components.OPENFGA:
		return openfga.NewController(options...), nil
	default:
		return nil, fmt.Errorf("unsupported policy language '%s'", cfg.PolicyLanguage)
	}
}

type authHandler struct {
	cfg        *config.Config
	logger     *slog.Logger
	controller pdp.Controller
}
