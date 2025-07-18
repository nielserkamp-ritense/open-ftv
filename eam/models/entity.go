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

// Entity represents the interface to work with the details of an entity.
//
// Entity is an immutable object and is by design safe for use in concurrent go-routines.
type Entity interface {
	UID() string              // retrieve the Unique ID (UID) of the entity.
	Type() string             // retrieve the Type of the entity (e,g, name-space).
	ID() string               // retrieve the ID of the entity (unique ID within the name-space).
	Attributes() AttributeSet // retrieve attributes of the entity.
	Parents() []string        // retrieve unique identifiers (UID) of parent entities.
	AddTags(tags ...string)   // associate tags with the entity.
	Tags() []string           // retrieve tags associated with the entity.
	HasTag(tag string) bool   // test if the entity contains a specific tag.
	MarshalJSON() ([]byte, error)
}

// NewEntity instanties a new standard entity.
func NewEntity(ns, id string, attrs AttributeSet, parents ...string) Entity {
	return &entity{
		uid:     EntityUID(ns, id),
		ns:      ns,
		id:      id,
		attrs:   attrs,
		parents: parents,
		tags:    make(map[string]struct{}),
	}
}

// UID implements the Entity interface.
func (e *entity) UID() string {
	return e.uid
}

// Type implements the Entity interface.
func (e *entity) Type() string {
	return e.ns
}

// ID implements the Entity interface.
func (e *entity) ID() string {
	return e.id
}

// Attributes implements the Entity interface.
func (e *entity) Attributes() AttributeSet {
	return e.attrs
}

// Parents implements the Entity interface.
func (e *entity) Parents() []string {
	return e.parents
}

// AddTags implements the Entity interface.
func (e *entity) AddTags(tags ...string) {
	for i := range tags {
		e.tags[tags[i]] = struct{}{}
	}
}

// Tags implements the Entity interface.
func (e *entity) Tags() []string {
	tags := make([]string, 0, len(e.tags))
	for k := range e.tags {
		tags = append(tags, k)
	}
	slices.Sort(tags)
	return tags
}

// HasTag implements the Entity interface.
func (e *entity) HasTag(tag string) bool {
	_, ok := e.tags[tag]
	return ok
}

// MarshalJSON implements the json.Marshaller interface.
func (e *entity) MarshalJSON() ([]byte, error) {
	return json.Marshal(entityJSON{
		Type:       e.ns,
		ID:         e.id,
		Attributes: e.attrs,
		Parents:    e.parents,
		Tags:       e.Tags(),
	})
}

// EntityToAttribute can be used to convert an entity into an attribute.
func EntityToAttribute(e Entity) Attribute {
	return NewAttribute(e.UID(), mapFromEntity(e))
}

func mapFromEntity(e Entity) map[string]any {
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

// EntityEqual returns true if the entities match exactly.
func EntityEqual(e1, e2 Entity) bool {
	return e1.UID() == e2.UID() &&
		reflect.DeepEqual(e1.Parents(), e2.Parents()) &&
		reflect.DeepEqual(e1.Tags(), e2.Tags()) &&
		AttributesEqual(e1.Attributes(), e2.Attributes())
}

type entity struct {
	uid     string              // unique identifier (ns::id).
	ns      string              // name-space.
	id      string              // unique identifier within the name-space.
	attrs   AttributeSet        // optional attributes.
	parents []string            // optional set of unique identifiers of parent entities.
	tags    map[string]struct{} // tags associated with the entity.
}

type entityJSON struct {
	Type       string       `json:"type,omitempty"`
	ID         string       `json:"id,omitempty"`
	Attributes AttributeSet `json:"attributes,omitempty"`
	Parents    []string     `json:"parents,omitempty"`
	Tags       []string     `json:"tags,omitempty"`
}
