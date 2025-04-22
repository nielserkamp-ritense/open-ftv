// Package pep contains functionality to implement a Policy Enforcement Point.
package pep

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// PEP represents the interface for a Policy Enforcement Point.
type PEP interface {
	PARCFromHTTP(uid uuid.UUID, req *models.HTTPRequest, attrs models.AttributeSet, e models.EntitySet) *models.PARC
	PARCFromRequest(req *models.Request, e models.EntitySet) *models.PARC
}

// New instantiates a new Policy Enforcement Point.
func New(ctx context.Context, logger *slog.Logger) PEP {
	if ctx == nil {
		ctx = context.Background()
	}
	return &pep{ctx: ctx, logger: logger}
}

type pep struct {
	ctx    context.Context
	logger *slog.Logger
}
