package models

// AddAttribute is the function signature for adding attributes to a set.
type AddAttribute func(*Attribute) (*Attribute, error)

// GetAttributeValue is the function signature for retrieving attribute values from a set.
type GetAttributeValue func(key string) any

// AddEntity is the function signature for adding entities to a set.
type AddEntity func(*Entity)

// GetEntity is the function signature for retrieving entities from a set.
type GetEntity func(uid string) *Entity

// AddRelation is the function signature for adding relations to a set.
type AddRelation func(*Relation)
