package server

import (
	"fmt"
	"log/slog"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	handlers "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cerbos-api"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/opa-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/openfga-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
)

// AuthHandler represents the interface for handling authorization requests.
type AuthHandler interface {
	Controller() pdp.Controller
	AuthZEN() *handlers.AuthZENAuthorizer // AuthZEN API.
	Authorizer() authorization.Authorizer // UI & bundle authorization.
}

func (s *service) newAuth(basePath string) AuthHandler {
	controller, err := s.newController()
	if controller == nil {
		s.logger.Error("failed to initialize pdp controller", "error", err)
		return nil
	}

	// authenticator, err2 := s.cfg.Authentication.NewAuthenticator(s.ctx, s.logger, func(uid string) (*models.Entity, uint64, error) {
	// 	return controller.GetPIP().GetEntity(uid)
	// })
	// if err2 != nil {
	// 	s.logger.Error("failed to initialize authenticator", "error", err2)
	// 	return nil
	// }

	authorizer, err3 := s.cfg.Authorization.NewAuthorizer(controller, nil)
	if err3 != nil {
		s.logger.Error("failed to initialize authorizer", "error", err3)
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

	var prefix string
	if s.cfg.AuthZENMethod != "" || s.cfg.AuthZENDomain != "" {
		prefix = fmt.Sprintf("%s:/%s%s", s.cfg.AuthZENMethod, s.cfg.AuthZENDomain, basePath)
	}

	zen := handlers.NewAuthHandlerZEN(s.logger, authLogger, controller, prefix).
		WithEvaluations()

	return &authHandler{
		logger:     s.logger,
		controller: controller,
		zen:        zen,
		authorizer: authorizer,
	}
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

// AuthZEN returns the AuthZEN API handler.
func (h *authHandler) AuthZEN() *handlers.AuthZENAuthorizer { return h.zen }

// Authorizer returns the authorizer.
func (h *authHandler) Authorizer() authorization.Authorizer { return h.authorizer }

type authHandler struct {
	logger     *slog.Logger
	controller pdp.Controller
	zen        *handlers.AuthZENAuthorizer
	authorizer authorization.Authorizer
}
