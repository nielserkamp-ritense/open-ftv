package authorization

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pep"
)

// Option represents the function signature to pass options for instantiating an authorization handler.
type Option func(a *auth)

// WithContext passes a context to the authorization handler.
func WithContext(ctx context.Context) Option {
	return func(a *auth) {
		a.ctx = ctx
	}
}

// WithLogger passes a logger to the authorization handler.
func WithLogger(logger *slog.Logger) Option {
	return func(a *auth) {
		a.log = logger
	}
}

// WithPDP passes a PDP to the authorization handler.
func WithPDP(pdp pdp.Controller) Option {
	return func(a *auth) {
		a.pdp = pdp
	}
}

// WithPEP passes a PEP to the authorization handler.
func WithPEP(pep pep.PEP) Option {
	return func(a *auth) {
		a.pep = pep
	}
}

// WithEntities passes the default set of entities to the authorization handler.
func WithEntities(entities models.EntitySet) Option {
	return func(a *auth) {
		a.entities = entities
	}
}

// WithAuthenticator passes an authentication handler to the authorization handler.
func WithAuthenticator(authenticator authentication.Authenticator) Option {
	return func(a *auth) {
		a.authenticator = authenticator
	}
}
