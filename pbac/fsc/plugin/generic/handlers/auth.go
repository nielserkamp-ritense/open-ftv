// Package handlers contains the HTTP request handlers for the FSC Auth plugin.
package handlers

import (
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
	AuthFSC(req *fiber.Ctx) error
	AuthZEN(req *fiber.Ctx) error
}

// New instantiates an authorization handler.
func New(cfg *config.Config, logger *slog.Logger, logboek ldv.LDV) AuthHandler {
	c, err := newController(cfg, logger, logboek)
	if c == nil {
		logger.Error("configuration error", "error", err)
		return nil
	}

	return &authHandler{cfg: cfg, logger: logger, controller: c}
}

func newController(cfg *config.Config, logger *slog.Logger, logboek ldv.LDV) (pdp.Controller, error) {
	switch components.LanguageFromString(cfg.PolicyLanguage) {
	case components.REGO:
		p := pip.New(pip.Config{
			Ctx:     nil,
			Store:   cfg.PipStore,
			Recurse: cfg.PipStoreRecurse,
			Logger:  logger,
		})
		return opa.NewController(p, cfg.PolicyStore, cfg.PolicyStoreRecurse, logger, logboek), nil

	case components.CERBOS:
		p := pip.New(pip.Config{
			Ctx:     nil,
			Store:   cfg.PipStore,
			Recurse: cfg.PipStoreRecurse,
			Logger:  logger,
		})
		return cerbos.NewController(p, cfg.PolicyStore, cfg.PolicyStoreRecurse, logger, logboek), nil

	case components.CEDAR:
		p := pip.New(pip.Config{
			Ctx:           nil,
			Store:         cfg.PipStore,
			Recurse:       cfg.PipStoreRecurse,
			Logger:        logger,
			NewAttributes: cedar.NewAttributeBuilder(logger),
			NewEntities:   cedar.NewEntityBuilder(logger),
		})
		return cedar.NewController(p, cfg.PolicyStore, cfg.PolicyStoreRecurse, logger, logboek), nil

	case components.OPENFGA:
		p := pip.New(pip.Config{
			Ctx:     nil,
			Store:   cfg.PipStore,
			Recurse: cfg.PipStoreRecurse,
			Logger:  logger,
		})
		return openfga.NewController(p, cfg.PolicyStore, cfg.PolicyStoreRecurse, logger, logboek), nil

	default:
		return nil, fmt.Errorf("unsupported policy language '%s'", cfg.PolicyLanguage)
	}
}

type authHandler struct {
	cfg        *config.Config
	logger     *slog.Logger
	controller pdp.Controller
}
