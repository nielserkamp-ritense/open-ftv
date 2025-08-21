package pap

import (
	"context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// LanguagePersister represents the interface for managing policy languages in a PAP persistence store.
type LanguagePersister interface {
	ListLanguages(ctx context.Context) (policies.Languages, error)
}

// TagPersister represents the interface for managing tags in a PAP persistence store.
type TagPersister interface {
	ListTags(ctx context.Context) (policies.Tags, error)
	ReplaceAllTags(tags []*policies.Tag, user string) error
	// CreateTag(ctx context.Context, user string, p *policies.Tag) (*policies.Tag, error)
	// ReadTag(ctx context.Context, id string) (*policies.Tag, uint64, error)
	// UpdateTag(ctx context.Context, user string, prev *policies.Tag, lastIndex uint64, p *policies.Tag) (*policies.Tag, error)
	// DeleteTag(ctx context.Context, user string, prev *policies.Tag, lastIndex uint64) (*policies.Tag, error)
}

// PolicyPersister represents the interface for managing policies in a PAP persistence store.
type PolicyPersister interface {
	CreatePolicy(ctx context.Context, user string, p *models.Policy) (*models.Policy, error)
	ReadPolicy(ctx context.Context, id string) (*models.Policy, uint64, error)
	UpdatePolicy(ctx context.Context, user string, prev *models.Policy, lastIndex uint64, p *models.Policy) (*models.Policy, error)
	DeletePolicy(ctx context.Context, user string, prev *models.Policy, lastIndex uint64) (*models.Policy, error)
	ListPolicies(ctx context.Context, language string) ([]*models.Policy, error)
}
