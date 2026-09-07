// Package server contains the HTTP request handlers for the app.
package server

import (
	"log/slog"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
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
	controller, p1, err := s.newSelfAuthzController()
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

// newSelfAuthzController builds the embedded PDP that authorizes this app's own HTTP API. s.pap must
// already be initialized: reusing it here means the embedded PDP enforces the very same
// policies the admin API manages, instead of a separate bundled copy.
func (s *Services) newSelfAuthzController() (pdp.Controller, *pip.PIP, error) {
	return config.BuildSelfAuthzController(s.ctx, s.logger, s.l, &s.cfg.PIP, &s.cfg.Cerbos).
		WithPAP(s.pap).
		Build()
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
// because pap.PAP does not expose the pool it holds internally.
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
