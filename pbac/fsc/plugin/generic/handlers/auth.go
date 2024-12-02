// Package handlers contains the HTTP request handlers for the FSC Auth plugin.
package handlers

import (
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/ldv"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/control"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/control/cedar"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/control/cerbos"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/control/opa"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/types"
)

// AuthHandler represents the interface for handling authorization requests.
type AuthHandler interface {
	AuthFSC(req *fiber.Ctx) error
	AuthZEN(req *fiber.Ctx) error
}

// New instantiates an authorization handler.
func New(cfg *config.Config, logger *slog.Logger, ldv ldv.LDV) AuthHandler {
	c, err := newController(cfg, logger, ldv)
	if c == nil {
		logger.Error("configuration error", "error", err)
		return nil
	}

	h := &authHandler{cfg: cfg, logger: logger, controller: c}
	return h
}

func newController(cfg *config.Config, logger *slog.Logger, ldv ldv.LDV) (control.Controller, error) {
	switch types.LanguageFromString(cfg.PolicyLanguage) {
	case types.REGO:
		p := pip.New(cfg.PipStore, cfg.PipStoreRecurse, logger, nil, nil)
		return opa.NewController(p, cfg.PolicyStore, cfg.PolicyStoreRecurse, logger, ldv), nil
	case types.CERBOS:
		p := pip.New(cfg.PipStore, cfg.PipStoreRecurse, logger, nil, nil)
		return cerbos.NewController(p, cfg.PolicyStore, cfg.PolicyStoreRecurse, logger, ldv), nil
	case types.CEDAR:
		p := pip.New(cfg.PipStore, cfg.PipStoreRecurse, logger, cedar.NewAttributeBuilder(logger), cedar.NewEntityBuilder(logger))
		return cedar.NewController(p, cfg.PolicyStore, cfg.PolicyStoreRecurse, logger, ldv), nil
	default:
		return nil, fmt.Errorf("unsupported policy language '%s'", cfg.PolicyLanguage)
	}
}

type authHandler struct {
	cfg        *config.Config
	logger     *slog.Logger
	controller control.Controller
}
