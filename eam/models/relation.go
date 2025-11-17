package models

import (
	"fmt"
	"reflect"
	"slices"
)

// Relation contains the details of a relationship between a subject and an object.
//
// Relation is an immutable object and is by design safe for use by concurrent go-routines.
type Relation struct {
	status      Status
	uid         string
	subject     *Entity
	predicate   *Entity
	object      *Entity
	title       string
	description string
	tags        map[string]struct{}
}

// NewRelationFromUID instantiates a new relation.
//
// The given subject-, predicate- and object-key are used to find the entities in the given set.
func NewRelationFromUID(subject, predicate, object string, entities *EntitySet) *Relation {
	s := entities.GetEntity(subject)
	p := entities.GetEntity(predicate)
	o := entities.GetEntity(object)
	return NewRelation(s, p, o)
}

// NewRelation instantiates a new relation.
func NewRelation(subject, predicate, object *Entity) *Relation {
	return &Relation{
		status:    StatusConcept,
		uid:       fmt.Sprintf("%s|%s|%s", getUID(subject), getUID(predicate), getUID(object)),
		subject:   subject,
		predicate: predicate,
		object:    object,
		tags:      make(map[string]struct{}),
	}
}

// WithStatus sets the status of the Relation.
func (r *Relation) WithStatus(status Status) *Relation {
	r.status = status
	return r
}

// WithTitle adds an optional title to the Relation.
func (r *Relation) WithTitle(title string) *Relation {
	r.title = title
	return r
}

// WithDescription adds an optional description to the Relation.
func (r *Relation) WithDescription(desc string) *Relation {
	r.description = desc
	return r
}

// WithTags annotates the Relation with the given tags.
func (r *Relation) WithTags(tags ...string) *Relation {
	for i := range tags {
		r.tags[tags[i]] = struct{}{}
	}
	return r
}

func getUID(r *Entity) string {
	if r == nil {
		return "?"
	}
	return r.UID()
}

// Status returns the current status of the Relation.
func (r *Relation) Status() Status {
	return r.status
}

// StatusName returns the current status of the Relation as a string.
func (r *Relation) StatusName() string {
	return r.status.String()
}

// UID returns the unique identifier of the Relation.
func (r *Relation) UID() string {
	return r.uid
}

// Subject returns the subject of the Relation.
func (r *Relation) Subject() *Entity {
	return r.subject
}

// Predicate returns the predicate of the Relation.
func (r *Relation) Predicate() *Entity {
	return r.predicate
}

// Object returns the object of the Relation.
func (r *Relation) Object() *Entity {
	return r.object
}

// Title returns the title of the Relation.
func (r *Relation) Title() string {
	return r.title
}

// Description returns the description of the Relation.
func (r *Relation) Description() string {
	return r.description
}

// Tags returns the tags for the Relation.
func (r *Relation) Tags() []string {
	tags := make([]string, 0, len(r.tags))
	for k := range r.tags {
		tags = append(tags, k)
	}
	slices.Sort(tags)
	return tags
}

// HasTag returns true if the Relations is annotated with the given tag.
func (r *Relation) HasTag(tag string) bool {
	_, ok := r.tags[tag]
	return ok
}

// Equals returns true if this Relation equals the other Relation.
func (r *Relation) Equals(other *Relation) bool {
	return r.status == other.status &&
		r.uid == other.uid &&
		r.title == other.title &&
		r.description == other.description &&
		r.subject.Equals(other.subject) &&
		r.predicate.Equals(other.predicate) &&
		r.object.Equals(other.object) &&
		reflect.DeepEqual(r.tags, other.tags)
}

// RelationToAttribute can be used to convert a relation into an attribute.
func RelationToAttribute(r *Relation) *Attribute {
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
