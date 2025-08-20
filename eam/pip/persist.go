package pip

import (
	"context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// AttributePersister represents the interface to manage persistent storage for attributes.
type AttributePersister interface {
	CreateAttribute(ctx context.Context, user string, a *models.Attribute) (*models.Attribute, error)
	ReadAttribute(ctx context.Context, key string) (*models.Attribute, uint64, error)
	UpdateAttribute(ctx context.Context, user string, prev *models.Attribute, lastIndex uint64, a *models.Attribute) (*models.Attribute, error)
	DeleteAttribute(ctx context.Context, user string, prev *models.Attribute, lastIndex uint64) (*models.Attribute, error)
	ListAttributes(ctx context.Context) ([]*models.Attribute, error)
}

// EntityPersister represents the interface to manage persistent storage for attributes.
type EntityPersister interface {
	CreateEntity(ctx context.Context, user string, a *models.Entity) (*models.Entity, error)
	ReadEntity(ctx context.Context, ns, id string) (*models.Entity, uint64, error)
	UpdateEntity(ctx context.Context, user string, prev *models.Entity, lastIndex uint64, a *models.Entity) (*models.Entity, error)
	DeleteEntity(ctx context.Context, user string, prev *models.Entity, lastIndex uint64) (*models.Entity, error)
	ListEntities(ctx context.Context) ([]*models.Entity, error)
}

// RelationPersister represents the interface to manage persistent storage for relations.
type RelationPersister interface {
	CreateRelation(ctx context.Context, user string, r *models.Relation) (*models.Relation, error)
	ReadRelation(ctx context.Context, id string) (*models.Relation, uint64, error)
	UpdateRelation(ctx context.Context, user string, prev *models.Relation, lastIndex uint64, r *models.Relation) (*models.Relation, error)
	DeleteRelation(ctx context.Context, user string, prev *models.Relation, lastIndex uint64) (*models.Relation, error)
	ListRelations(ctx context.Context) ([]*models.Relation, error)
}
