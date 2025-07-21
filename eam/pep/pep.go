// Package pep contains functionality to implement a Policy Enforcement Point.
package pep

import (
	"context"
	"log/slog"
)

// PEP contains the interface for a Policy Enforcement Point.
type PEP struct {
	ctx    context.Context
	logger *slog.Logger
}

// New instantiates a new Policy Enforcement Point.
func New(ctx context.Context, logger *slog.Logger) *PEP {
	if ctx == nil {
		ctx = context.Background()
	}
	return &PEP{ctx: ctx, logger: logger}
}
