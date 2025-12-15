package pip

import (
	"context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

// AttributePersister represents the interface to manage persistent storage for attributes.
type AttributePersister interface {
	CreateAttribute(ctx context.Context, a *models.Attribute) (*models.Attribute, error)
	ReadAttribute(ctx context.Context, key string) (*models.Attribute, uint64, error)
	ReadAttributeAudit(ctx context.Context, key string) ([]oas.AuditEntry, error)
	ReadAttributeDeployments(ctx context.Context, key string) ([]oas.UsageData, error)
	ReadAttributeVersions(ctx context.Context, id string) (oas.AttributeVersions, error)
	ReadAttributeVersion(ctx context.Context, id string, version int) (*oas.AttributeVersion, error)
	UpdateAttribute(ctx context.Context, prev *models.Attribute, lastIndex uint64, a *models.Attribute) (*models.Attribute, error)
	DeleteAttribute(ctx context.Context, prev *models.Attribute, lastIndex uint64) (*models.Attribute, error)
	ListAttributes(ctx context.Context) ([]*models.Attribute, error)
}

// EntityPersister represents the interface to manage persistent storage for attributes.
type EntityPersister interface {
	CreateEntity(ctx context.Context, a *models.Entity) (*models.Entity, error)
	ReadEntity(ctx context.Context, ns, id string) (*models.Entity, uint64, error)
	ReadEntityAudit(ctx context.Context, ns, id string) ([]oas.AuditEntry, error)
	ReadEntityDeployments(ctx context.Context, ns, id string) ([]oas.UsageData, error)
	ReadEntityVersions(ctx context.Context, ns, id string) (oas.EntityVersions, error)
	ReadEntityVersion(ctx context.Context, ns, id string, version int) (*oas.EntityVersion, error)
	UpdateEntity(ctx context.Context, prev *models.Entity, lastIndex uint64, a *models.Entity) (*models.Entity, error)
	DeleteEntity(ctx context.Context, prev *models.Entity, lastIndex uint64) (*models.Entity, error)
	ListEntities(ctx context.Context) ([]*models.Entity, error)
}

// RelationPersister represents the interface to manage persistent storage for relations.
type RelationPersister interface {
	CreateRelation(ctx context.Context, r *models.Relation) (*models.Relation, error)
	ReadRelation(ctx context.Context, id string) (*models.Relation, uint64, error)
	ReadRelationAudit(ctx context.Context, id string) ([]oas.AuditEntry, error)
	ReadRelationDeployments(ctx context.Context, id string) ([]oas.UsageData, error)
	ReadRelationVersions(ctx context.Context, id string) (oas.RelationVersions, error)
	ReadRelationVersion(ctx context.Context, id string, version int) (*oas.RelationVersion, error)
	UpdateRelation(ctx context.Context, prev *models.Relation, lastIndex uint64, r *models.Relation) (*models.Relation, error)
	DeleteRelation(ctx context.Context, prev *models.Relation, lastIndex uint64) (*models.Relation, error)
	ListRelations(ctx context.Context) ([]*models.Relation, error)
}
