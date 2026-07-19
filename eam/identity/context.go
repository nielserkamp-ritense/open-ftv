package identity

import "context"

type ctxKey struct{}

// WithContext attaches p to ctx, for the rare case where a function's signature can't carry a
// Principal directly (e.g. a generic Exec(ctx, query, params) in Postgres).
func WithContext(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, ctxKey{}, p)
}

// FromContext retrieves the Principal attached via WithContext, if any.
func FromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(ctxKey{}).(Principal)
	return p, ok
}
