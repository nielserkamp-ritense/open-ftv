package pap

import (
	"context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// LanguagePersister represents the interface for managing policy languages in a PAP persistence store.
type LanguagePersister interface {
	ListLanguages(ctx context.Context) (oas.Languages, error)
}

// TagPersister represents the interface for managing tags in a PAP persistence store.
type TagPersister interface {
	ListTags(ctx context.Context) (oas.Tags, error)
	CreateTag(ctx context.Context, t *oas.Tag) (*oas.Tag, error)
	ReadTag(ctx context.Context, id string) (*oas.Tag, uint64, error)
	UpdateTag(ctx context.Context, prev *oas.Tag, lastIndex uint64, t *oas.Tag) (*oas.Tag, error)
	DeleteTag(ctx context.Context, prev *oas.Tag, lastIndex uint64) (*oas.Tag, error)
	ReplaceAllTags(tags []*oas.Tag, user string) error
}

// PolicyPersister represents the interface for managing policies in a PAP persistence store.
type PolicyPersister interface {
	CreatePolicy(ctx context.Context, p *models.Policy) (*models.Policy, error)
	ReadPolicy(ctx context.Context, id string) (*models.Policy, uint64, error)
	ReadPolicyAudit(ctx context.Context, id string) ([]oas.AuditEntry, error)
	ReadPolicyDeployments(ctx context.Context, id string) ([]oas.UsageData, error)
	UpdatePolicy(ctx context.Context, prev *models.Policy, lastIndex uint64, p *models.Policy) (*models.Policy, error)
	DeletePolicy(ctx context.Context, prev *models.Policy, lastIndex uint64) (*models.Policy, error)
	ListPolicies(ctx context.Context, language string) ([]*models.Policy, error)
}
