package authentication

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Option represents the function signature to pass options for instantiating an authentication handler.
type Option func(b *base)

// WithContext passes a context to the authentication handler.
func WithContext(ctx context.Context) Option {
	return func(b *base) {
		b.ctx = ctx
	}
}

// WithLogger passes a logger to the authentication handler.
func WithLogger(logger *slog.Logger) Option {
	return func(b *base) {
		b.log = logger
	}
}

// WithEntityGetter passes an entity retrieval function to the authentication handler.
func WithEntityGetter(getter models.GetEntity) Option {
	return func(b *base) {
		b.getEntity = getter
	}
}
