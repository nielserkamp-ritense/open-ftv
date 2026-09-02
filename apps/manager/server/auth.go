// Package server contains the HTTP request handlers for the app.
package server

import (
	"fmt"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/principals"
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

	// Record principals whenever this manager is postgres-backed. The standalone PAP and PIP do
	// the same for their own writes: they share this schema, so created_by there is bound by the
	// same foreign key (see their principalRecorder). The PDP needs none -- its POST endpoints are
	// AuthZEN evaluation and the bundle receiver, and its postgres holds only the decision log.
	var extra []authorization.Option
	if s.db != nil {
		extra = append(extra, authorization.WithPrincipalRecorder(principals.NewDBWithPool(s.db)))
	}

	authorizer, err3 := s.cfg.NewAuthorizer(controller, authenticator, extra...)
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

func (s *Services) newSelfAuthzController() (pdp.Controller, *pip.PIP, error) {
	// Secured mode (fail-closed) requires OIDC: without a validated token no principal is
	// derived, so fail-closed would deny every request. Fail fast with a clear error rather
	// than booting a manager that silently rejects everything.
	secured := s.cfg.FailClosedOnEmpty

	opt, err := pepOIDCOption(s.ctx, s.cfg.OIDC)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build OIDC validation: %w", err)
	}

	if opt == nil {
		if secured {
			return nil, nil, fmt.Errorf("OIDC is required in fail-closed mode: set the OIDC_JWKS_URL (MANAGER_OIDC_JWKS_URL) so bearer tokens yield a principal, otherwise all requests are denied")
		}

		s.logger.Warn("manager: OIDC not configured; bearer tokens will not be validated and yield no principal")
	}

	var pepOpts []pep.Option
	if opt != nil {
		pepOpts = append(pepOpts, opt)
	}

	return config.BuildSelfAuthzController(s.ctx, s.logger, s.l, &s.cfg.PIP, &s.cfg.Cerbos).
		WithPAP(s.pap).
		WithOptions(pdp.WithPEP(pep.New(s.ctx, s.logger, pepOpts...))).
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
