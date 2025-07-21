package models

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/goccy/go-json"
)

// EntityUID formats the unique identifier (UID) for an entity.
func EntityUID(ns, id string) string {
	return fmt.Sprintf("%s::%s", ns, id)
}

// Entity contains the details of an entity.
//
// Entity is an immutable object and is by design safe for use in concurrent go-routines.
type Entity struct {
	uid     string              // unique identifier (ns::id).
	ns      string              // name-space.
	id      string              // unique identifier within the name-space.
	attrs   *AttributeSet       // optional attributes.
	parents []string            // optional set of unique identifiers of parent entities.
	tags    map[string]struct{} // tags associated with the entity.
}

// NewEntity instanties a new standard entity.
func NewEntity(ns, id string, attrs *AttributeSet, parents ...string) *Entity {
	return &Entity{
		uid:     EntityUID(ns, id),
		ns:      ns,
		id:      id,
		attrs:   attrs,
		parents: parents,
		tags:    make(map[string]struct{}),
	}
}

// UID returns the unique identifier (Type() + ID()) for the Entity.
func (e *Entity) UID() string {
	return e.uid
}

// Type returns the type of Entity.
func (e *Entity) Type() string {
	return e.ns
}

// ID returns the identifier for the Entity.
func (e *Entity) ID() string {
	return e.id
}

// Attributes returns the attributes for the Entity.
func (e *Entity) Attributes() *AttributeSet {
	return e.attrs
}

// Parents returns the unique parents for the Entity.
func (e *Entity) Parents() []string {
	return e.parents
}

// AddTags associates the given tags with the Entity.
func (e *Entity) AddTags(tags ...string) {
	for i := range tags {
		e.tags[tags[i]] = struct{}{}
	}
}

// Tags returns the tags for the Entity.
func (e *Entity) Tags() []string {
	tags := make([]string, 0, len(e.tags))
	for k := range e.tags {
		tags = append(tags, k)
	}
	slices.Sort(tags)
	return tags
}

// HasTag returns true if the Entity is associated with the given tag.
func (e *Entity) HasTag(tag string) bool {
	_, ok := e.tags[tag]
	return ok
}

// MarshalJSON implements the json.Marshaler interface.
func (e *Entity) MarshalJSON() ([]byte, error) {
	return json.Marshal(entityJSON{
		Type:       e.ns,
		ID:         e.id,
		Attributes: e.attrs,
		Parents:    e.parents,
		Tags:       e.Tags(),
	})
}

// EntityToAttribute can be used to convert an entity into an attribute.
func EntityToAttribute(e *Entity) *Attribute {
	return NewAttribute(e.UID(), mapFromEntity(e))
}

func mapFromEntity(e *Entity) map[string]any {
	m := map[string]any{"type": e.Type(), "id": e.ID()}

	if attr := MapFromAttributes(e.Attributes()); len(attr) > 0 {
		m["attributes"] = attr
	}
	if parents := e.Parents(); len(parents) > 0 {
		m["parents"] = parents
	}
	if tags := e.Tags(); len(tags) > 0 {
		m["tags"] = tags
	}

	return m
}

// Equals returns true if this Entity matches exactly with the other entity.
func (e *Entity) Equals(other *Entity) bool {
	return e.UID() == other.UID() &&
		reflect.DeepEqual(e.Parents(), other.Parents()) &&
		reflect.DeepEqual(e.Tags(), other.Tags()) &&
		e.Attributes().Equals(other.Attributes())
}

type entityJSON struct {
	Type       string        `json:"type,omitempty"`
	ID         string        `json:"id,omitempty"`
	Attributes *AttributeSet `json:"attributes,omitempty"`
	Parents    []string      `json:"parents,omitempty"`
	Tags       []string      `json:"tags,omitempty"`
}
