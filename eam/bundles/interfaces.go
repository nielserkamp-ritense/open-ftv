package bundles

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"

// PolicyLister represents the interface for retrieving a list of policies.
type PolicyLister interface {
	Iterate(f models.PolicyIterator)
}

// AttributeLister represents the interface for retrieving a list of attributes.
type AttributeLister interface {
	IterateAttributes(f models.AttributeIterator)
}

// EntityLister represents the interface for retrieving a list of entities.
type EntityLister interface {
	IterateEntities(f models.EntityIterator)
}

// RelationLister represents the interface for retrieving a list of relations.
type RelationLister interface {
	IterateRelations(f models.RelationIterator)
}
