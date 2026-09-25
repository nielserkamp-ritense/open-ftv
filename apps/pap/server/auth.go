// Package server contains the HTTP request handlers for the app.
package server

import (
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	cerbos_api "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cerbos-api"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/principals"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/selfauthz"
)

// AuthHandler represents the interface for handling authorization requests.
type AuthHandler interface {
	Controller() pdp.Controller
	Authorizer() authorization.RequestAuthorizer
}

func (s *Services) newAuth() AuthHandler {
	controller, p1, err := s.newSelfAuthzController()
	if controller == nil {
		s.logger.Error("failed to initialize EAM controller", "error", err)
		return nil
	}

	authenticator, err2 := authentication.New(s.ctx, s.logger, s.cfg.Type, p1.GetEntity)
	if err2 != nil {
		s.logger.Error("failed to initialize authenticator", "error", err2)
		return nil
	}

	// The management-plane services decide on the PAP's own role table (ADR 0006).
	opts := append([]authorization.Option{authorization.WithPermissions(authorization.PAPPermissions())}, s.principalRecorder()...)

	authorizer, err3 := authorization.NewForController(controller, authenticator, s.cfg.Authenticate, s.cfg.FailClosedOnEmpty, opts...)
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
	ip, err := pip.NewSelfAuthz(s.ctx, s.logger, s.cfg.PIP.Store, s.cfg.PIP.StoreRecurse, s.cfg.PullConfigs)
	if err != nil {
		return nil, nil, err
	}

	controller, err := selfauthz.NewController(s.ctx, s.logger, s.l, ip, s.pap, (*cerbos_api.Config)(&s.cfg.Cerbos))

	return controller, ip, err
}

// Controller returns the PDP controller.
func (h *authHandler) Controller() pdp.Controller { return h.controller }

// Authorizer returns the authorizer.
func (h *authHandler) Authorizer() authorization.RequestAuthorizer { return h.authorizer }

type authHandler struct {
	logger        *slog.Logger
	controller    pdp.Controller
	authenticator authentication.Authenticator
	authorizer    authorization.RequestAuthorizer
}

// principalRecorder returns the principal store on the shared connection pool.
//
// The standalone PAP writes created_by/updated_by into the same schema the manager uses, where
// migration 00014 makes those columns foreign keys to principal. Without recording the caller,
// every create and update here would fail on a foreign-key violation.
func (s *Services) principalRecorder() []authorization.Option {
	if s.db == nil {
		return nil
	}

	return []authorization.Option{
		authorization.WithPrincipalRecorder(principals.NewDBWithPool(s.db)),
	}
}
