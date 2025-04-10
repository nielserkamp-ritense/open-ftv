// Package server contains the HTTP request handlers for the app.
package server

import (
	"fmt"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/cedar"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/cerbos"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/opa"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/openfga"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// AuthHandler represents the interface for handling authorization requests.
type AuthHandler interface {
	Controller() pdp.Controller
	Authorizer() authorization.Authorizer
}

func (s *service) newAuth() AuthHandler {
	controller, err := s.newController()
	if controller == nil {
		s.logger.Error("failed to initialize EAM controller", "error", err)
		return nil
	}

	authenticator, err2 := s.cfg.Authentication.NewAuthenticator(controller)
	if err2 != nil {
		s.logger.Error("failed to initialize authenticator", "error", err2)
		return nil
	}

	authorizer, err3 := s.cfg.Authorization.NewAuthorizer(controller, authenticator)
	if err3 != nil {
		s.logger.Error("failed to initialize authorizer", "error", err3)
		return nil
	}

	return &authHandler{
		logger:        s.logger,
		controller:    controller,
		authenticator: authenticator,
		authorizer:    authorizer,
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
		return cedar.NewController(options...), nil
	case models.REGO:
		return opa.NewController(options...), nil
	case models.OPENFGA:
		return openfga.NewController(options...), nil
	case models.CERBOS:
		cerbosCFG := cerbos.Config{Addr1: s.cfg.Cerbos.Address, Addr2: s.cfg.Cerbos.AdminAddress, CA: s.cfg.Cerbos.CA, User: s.cfg.Cerbos.User, Pswd: s.cfg.Cerbos.Pswd}
		return cerbos.NewController(cerbosCFG, options...), nil
	default:
		return nil, fmt.Errorf("unsupported policy language '%s'", s.cfg.PAP.Language)
	}
}

// Controller returns the PDP controller.
func (h *authHandler) Controller() pdp.Controller { return h.controller }

// Authorizer returns the authorizer.
func (h *authHandler) Authorizer() authorization.Authorizer { return h.authorizer }

type authHandler struct {
	logger        *slog.Logger
	controller    pdp.Controller
	authenticator authentication.Authenticator
	authorizer    authorization.Authorizer
}
