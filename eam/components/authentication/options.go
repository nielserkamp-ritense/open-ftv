package authentication

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// Option represents the function signature to pass options for instantiating an authentication handler.
type Option func(a *bcryptAuth)

// WithContext passes a context to the authentication handler.
func WithContext(ctx context.Context) Option {
	return func(a *bcryptAuth) {
		a.ctx = ctx
	}
}

// WithLogger passes a logger to the authentication handler.
func WithLogger(logger *slog.Logger) Option {
	return func(a *bcryptAuth) {
		a.log = logger
	}
}

// WithEntities passes the default set of entities to the authentication handler.
func WithEntities(entities models.EntitySet) Option {
	return func(a *bcryptAuth) {
		a.entities = entities
	}
}
