package pep

import (
	"context"
	"log/slog"
)

type PEP struct {
	ctx    context.Context
	logger *slog.Logger
	jwt    *JWTConfig
}

func New(ctx context.Context, logger *slog.Logger, opts ...Option) *PEP {
	if ctx == nil {
		ctx = context.Background()
	}
	p := &PEP{ctx: ctx, logger: logger}
	for _, o := range opts {
		o(p)
	}
	return p
}
