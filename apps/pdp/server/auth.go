package server

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"

	handlers "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pdp/cedar-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pdp/cerbos-api"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pdp/opa-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pdp/openfga-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pep"
)

// AuthHandler represents the interface for handling authorization requests.
type AuthHandler interface {
	Controller() pdp.Controller
	AuthZEN(req *fiber.Ctx) error
}

func (s *service) newAuth() AuthHandler {
	controller, err := s.newController()
	if controller == nil {
		s.logger.Error("failed to initialize pdp controller", "error", err)
		return nil
	}

	var authLogger authlog.Logger
	if s.cfg.OpenSearch.Index != "" {
		authLogger, err = authlog.NewOpenSearch(s.cfg.OpenSearch.Index, s.cfg.OpenSearch.User, s.cfg.OpenSearch.Pswd, strings.Split(s.cfg.OpenSearch.Endpoints, ",")...)
		if err != nil {
			s.logger.Error("failed to initialize authlog", "index", s.cfg.OpenSearch.Index, "user", s.cfg.OpenSearch.User, "endpoints", s.cfg.OpenSearch.Endpoints, "error", err)
			return nil
		}
	}

	zen := handlers.NewAuthHandlerZEN(s.logger, authLogger, controller)

	return &authHandler{logger: s.logger, controller: controller, zen: zen}
}

func (s *service) newController() (pdp.Controller, error) {
	ep := pep.New(s.ctx, s.logger)

	ip, err := s.cfg.PIP.NewPIP(s.ctx, s.logger, s.l)
	if err != nil {
		return nil, err
	}

	ap, err2 := s.cfg.PAP.NewPAP(s.ctx, s.logger)
	if err2 != nil {
		return nil, err2
	}

	options := []pdp.Option{pdp.WithContext(s.ctx), pdp.WithLogger(s.logger), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap)}

	switch s.l {
	case models.CEDAR:
		return cedar_embedded.NewController(options...), nil
	case models.REGO:
		return opa_embedded.NewController(options...), nil
	case models.OPENFGA:
		return openfga_embedded.NewController(options...), nil
	case models.CERBOS:
		cerbosCFG := cerbos_api.Config{Addr1: s.cfg.Cerbos.Address, Addr2: s.cfg.Cerbos.AdminAddress, CA: s.cfg.Cerbos.CA, User: s.cfg.Cerbos.User, Pswd: s.cfg.Cerbos.Pswd}
		return cerbos_api.NewController(cerbosCFG, options...), nil
	default:
		return nil, fmt.Errorf("unsupported policy language '%s'", s.cfg.PAP.Language)
	}
}

// Controller returns the PDP controller.
func (h *authHandler) Controller() pdp.Controller { return h.controller }

// AuthZEN authorizes an AuthZEN authorization request.
func (h *authHandler) AuthZEN(req *fiber.Ctx) error { return h.zen.Authorize(req) }

type authHandler struct {
	logger     *slog.Logger
	controller pdp.Controller
	zen        handlers.AuthZENAuthorizer
}
