package models

import (
	"fmt"
	"slices"
)

// Relation represents the details of a relationship between a subject and an object.
//
// Relation is an immutable object and is by design safe for use by concurrent go-routines.
type Relation interface {
	UID() string            // return the unique identifier of this relation.
	Subject() Entity        // returns the subject of the relation.
	Predicate() Entity      // returns the predicate of the relation.
	Object() Entity         // returns the object of the relation.
	AddTags(tags ...string) // associate tags with the relation.
	Tags() []string         // retrieve tags associated with the relation.
	HasTag(tag string) bool // test if the relation contains a specific tag.
}

// NewRelationFromUID instantiates a new relation.
//
// The given subject-, predicate- and object-key are used to find the entities in the given set.
func NewRelationFromUID(subject, predicate, object string, entities EntitySet) Relation {
	s := entities.GetEntity(subject)
	p := entities.GetEntity(predicate)
	o := entities.GetEntity(object)
	return NewRelation(s, p, o)
}

// NewRelation instantiates a new relation.
func NewRelation(subject, predicate, object Entity) Relation {
	return &relation{
		uid:       fmt.Sprintf("%s|%s|%s", getUID(subject), getUID(predicate), getUID(object)),
		subject:   subject,
		predicate: predicate,
		object:    object,
		tags:      make(map[string]struct{}),
	}
}

func getUID(r Entity) string {
	if r == nil {
		return "?"
	}
	return r.UID()
}

// UID implements the Relation interface.
func (r *relation) UID() string {
	return r.uid
}

// Subject implements the Relation interface.
func (r *relation) Subject() Entity {
	return r.subject
}

// Predicate implements the Relation interface.
func (r *relation) Predicate() Entity {
	return r.predicate
}

// Object implements the Relation interface.
func (r *relation) Object() Entity {
	return r.object
}

// AddTags implements the Relation interface.
func (r *relation) AddTags(tags ...string) {
	for i := range tags {
		r.tags[tags[i]] = struct{}{}
	}
}

// Tags implements the Relation interface.
func (r *relation) Tags() []string {
	tags := make([]string, 0, len(r.tags))
	for k := range r.tags {
		tags = append(tags, k)
	}
	slices.Sort(tags)
	return tags
}

// HasTag implements the Relation interface.
func (r *relation) HasTag(tag string) bool {
	_, ok := r.tags[tag]
	return ok
}

// RelationToAttribute can be used to convert a relation into an attribute.
func RelationToAttribute(r Relation) Attribute {
	s := mapFromEntity(r.Subject())
	p := mapFromEntity(r.Predicate())
	o := mapFromEntity(r.Object())

	m := map[string]any{
		"subject":   s,
		"predicate": p,
		"object":    o,
	}

	return NewAttribute(r.UID(), m)
}

type relation struct {
	uid       string
	subject   Entity
	predicate Entity
	object    Entity
	tags      map[string]struct{}
}
