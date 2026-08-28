// Package server contains the HTTP request handlers for the app.
package server

import (
	"fmt"
	"log/slog"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	cedar_embedded "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	cerbos_api "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cerbos-api"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	opa_embedded "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/opa-embedded"
	openfga_embedded "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/openfga-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/principals"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

// AuthHandler represents the interface for handling authorization requests.
type AuthHandler interface {
	Controller() pdp.Controller
	Authorizer() authorization.Authorizer
}

func (s *Services) newAuth() AuthHandler {
	controller, p1, err := s.newController()
	if controller == nil {
		s.logger.Error("failed to initialize EAM controller", "error", err)
		return nil
	}

	authenticator, err2 := s.cfg.NewAuthenticator(s.ctx, s.logger, p1.GetEntity)
	if err2 != nil {
		s.logger.Error("failed to initialize authenticator", "error", err2)
		return nil
	}

	authorizer, err3 := s.cfg.NewAuthorizer(controller, authenticator, s.principalRecorder()...)
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

func (s *Services) newController() (pdp.Controller, *pip.PIP, error) {
	ep := pep.New(s.ctx, s.logger)

	ip, err := s.cfg.NewSelfAuthzPIP(s.ctx, s.logger, s.l)
	if err != nil {
		return nil, nil, err
	}

	ap, err2 := s.cfg.NewSelfAuthzPAP(s.ctx, s.logger)
	if err2 != nil {
		return nil, nil, err2
	}

	options := []pdp.Option{pdp.WithContext(s.ctx), pdp.WithLogger(s.logger), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap)}

	switch s.l {
	case models.CEDAR:
		return cedar_embedded.NewController(options...), ip, nil
	case models.REGO:
		return opa_embedded.NewController(options...), ip, nil
	case models.OPENFGA:
		return openfga_embedded.NewController(options...), ip, nil
	case models.CERBOS:
		cerbosCFG := cerbos_api.Config{Addr1: s.cfg.Address, Addr2: s.cfg.AdminAddress, CA: s.cfg.CA, User: s.cfg.User, Pswd: s.cfg.Pswd}
		return cerbos_api.NewController(cerbosCFG, options...), ip, nil
	default:
		return nil, nil, fmt.Errorf("unsupported policy language '%s'", s.cfg.Language)
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

// principalRecorder returns the principal store when this app is postgres-backed.
//
// The standalone PAP writes created_by/updated_by into the same schema the manager uses, where
// migration 00014 makes those columns foreign keys to principal. Without recording the caller,
// every create and update here would fail on a foreign-key violation. It opens its own pool
// because the PAP store is built after the authorizer.
func (s *Services) principalRecorder() []authorization.Option {
	if !strings.EqualFold(s.cfg.Persist.Type, "postgres") {
		return nil
	}

	db, err := postgresql.New(s.ctx, s.cfg.PgURL, s.cfg.PgMaxLife, s.cfg.PgMaxConn)
	if err != nil {
		s.logger.Error("failed to connect the principal store; writes will fail on the created_by foreign key", "error", err)
		return nil
	}

	return []authorization.Option{authorization.WithPrincipalRecorder(principals.NewDBWithPool(db))}
}
