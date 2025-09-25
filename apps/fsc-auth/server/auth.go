// Package server contains the HTTP request handlers for the FSC Auth plugin.
package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	handlers "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cerbos-api"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/opa-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/openfga-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"

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
	controller, err := newController(ctx, cfg, logger)
	if controller == nil {
		logger.Error("failed to initialize pdp controller", "error", err)
		return nil
	}

	fsc := handlers.NewAuthHandlerFSC(logger, controller)

	prefix := fmt.Sprintf("http://localhost:%d%s%s", cfg.Port, handlers.PathAuthZEN, handlers.PathV1)
	zen := handlers.NewAuthHandlerZEN(logger, nil, controller, prefix)

	return &authHandler{logger: logger, controller: controller, fsc: fsc, zen: zen}
}

func newController(ctx context.Context, cfg *config.Config, logger *slog.Logger) (pdp.Controller, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	l := models.LanguageFromString(cfg.PAP.Language)

	ep := pep.New(ctx, logger)

	pipOpts := []pip.Option{pip.WithFileStore(cfg.PIP.Store, cfg.PIP.StoreRecurse), pip.WithPullConfigs(cfg.PIP.PullConfigs)}
	ip := pip.New(ctx, logger, pipOpts...)

	papOpts := []pap.Option{pap.WithLanguage(l.Language()), pap.WithFileStore(cfg.PAP.Store, cfg.PAP.StoreRecurse)}
	ap := pap.New(ctx, logger, papOpts...)

	pdpOpts := []pdp.Option{pdp.WithContext(ctx), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger)}
	if cfg.RequestMappings != "" {
		pdpOpts = append(pdpOpts, pdp.WithMappings(mapping.MappingsFromConfig(cfg.RequestMappings)...))
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
func (h *authHandler) AuthFSC(req *fiber.Ctx) error { return h.fsc.Evaluation(req) }

// AuthZEN authorizes an AuthZEN authorization request.
func (h *authHandler) AuthZEN(req *fiber.Ctx) error { return h.zen.Evaluation(req) }

type authHandler struct {
	logger     *slog.Logger
	controller pdp.Controller
	fsc        handlers.FSCAuthorizer
	zen        *handlers.AuthZENAuthorizer
}
