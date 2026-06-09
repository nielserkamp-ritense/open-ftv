// Package server contains the HTTP request handlers for the app.
package server

import (
	"fmt"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cerbos-api"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/opa-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/openfga-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
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

	authenticator, err2 := s.cfg.Authentication.NewAuthenticator(s.ctx, s.logger, p1.GetEntity)
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

func (s *Services) newController() (pdp.Controller, *pip.PIP, error) {
	// Secured mode (fail-closed) requires OIDC: without a validated token no principal is
	// derived, so fail-closed would deny every request. Fail fast with a clear error rather
	// than booting a manager that silently rejects everything.
	secured := s.cfg.Authorization.FailClosedOnEmpty

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

	ep := pep.New(s.ctx, s.logger, pepOpts...)

	ip, err := s.cfg.PIP.NewPIP(s.ctx, s.logger, s.l)
	if err != nil {
		return nil, nil, err
	}

	// Share the single, already-initialized PAP (postgres-backed when configured)
	// so the embedded PDP enforces from the same policy store the UI manages.
	options := []pdp.Option{pdp.WithContext(s.ctx), pdp.WithLogger(s.logger), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(s.pap)}

	switch s.l {
	case models.CEDAR:
		return cedar_embedded.NewController(options...), ip, nil
	case models.REGO:
		return opa_embedded.NewController(options...), ip, nil
	case models.OPENFGA:
		return openfga_embedded.NewController(options...), ip, nil
	case models.CERBOS:
		cerbosCFG := cerbos_api.Config{Addr1: s.cfg.Cerbos.Address, Addr2: s.cfg.Cerbos.AdminAddress, CA: s.cfg.Cerbos.CA, User: s.cfg.Cerbos.User, Pswd: s.cfg.Cerbos.Pswd}
		return cerbos_api.NewController(cerbosCFG, options...), ip, nil
	default:
		return nil, nil, fmt.Errorf("unsupported policy language '%s'", s.cfg.PAP.Language)
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
