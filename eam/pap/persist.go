package pap

import (
	"context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// PolicyPersister represents the interface for a PAP persistence store.
type PolicyPersister interface {
	CreatePolicy(ctx context.Context, user string, p *models.Policy) (*models.Policy, error)
	ReadPolicy(ctx context.Context, id string) (*models.Policy, uint64, error)
	UpdatePolicy(ctx context.Context, user string, prev *models.Policy, lastIndex uint64, p *models.Policy) (*models.Policy, error)
	DeletePolicy(ctx context.Context, user string, prev *models.Policy, lastIndex uint64) (*models.Policy, error)
	ListPolicies(ctx context.Context, language string) ([]*models.Policy, error)
}
