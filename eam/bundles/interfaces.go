package bundles

import (
	"context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Persister represents the interface to manage deployments.
type Persister interface {
	Generate(ctx context.Context, title, description, user string) (*Deployment, error)
	Advance() (*Deployment, error)
	Fail(msg string) (*Deployment, error)
	LastDeployment(ctx context.Context) (*Deployment, error)
	ReadDeployment(ctx context.Context, version uint64) (*Deployment, error)
	ListDeployments(ctx context.Context) ([]*Deployment, error)
	CreateBundleAudit(ctx context.Context, principal identity.Principal, version uint64, cfg *Config, bundle *Bundle) error
}

// PolicyLister represents the interface for retrieving a list of policies.
type PolicyLister interface {
	Iterate(f models.PolicyIterator)
}

// PolicyCreator represents the interface to store a new policy.
type PolicyCreator interface {
	Create(in *models.Policy, user identity.Principal) (out *models.Policy, err error)
}

// AttributeLister represents the interface for retrieving a list of attributes.
type AttributeLister interface {
	IterateAttributes(f models.AttributeIterator)
}

// AttributeCreator represents the interface to store a new attribute.
type AttributeCreator interface {
	AddAttribute(in *models.Attribute) (*models.Attribute, error)
}

// EntityLister represents the interface for retrieving a list of entities.
type EntityLister interface {
	IterateEntities(f models.EntityIterator)
}

// EntityCreator represents the interface to store a new entity.
type EntityCreator interface {
	AddEntity(entity *models.Entity) (*models.Entity, error)
}

// RelationLister represents the interface for retrieving a list of relations.
type RelationLister interface {
	IterateRelations(f models.RelationIterator)
}

// RelationCreator represents the interface to store a new relation.
type RelationCreator interface {
	AddRelation(rel *models.Relation) (*models.Relation, error)
}

// PolicyHandler represents the interface to work with policies (PAP).
type PolicyHandler interface {
	PolicyLister
	PolicyCreator
}

// DataHandler represents the interface to work with attributes, entities and relations (PIP).
type DataHandler interface {
	AttributeLister
	AttributeCreator
	EntityLister
	EntityCreator
	RelationLister
	RelationCreator
}
