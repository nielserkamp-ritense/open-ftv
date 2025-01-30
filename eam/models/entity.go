package models

import (
	"fmt"

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

// MarshalJSON implements the json.Marshaller interface.
func (e *entity) MarshalJSON() ([]byte, error) {
	return json.Marshal(entityJSON{
		Type:       e.ns,
		ID:         e.id,
		Attributes: e.attrs,
		Parents:    e.parents,
	})
}

type entity struct {
	uid     string       // unique identifier (ns::id).
	ns      string       // name-space.
	id      string       // unique identifier within the name-space.
	attrs   AttributeSet // optional attributes.
	parents []string     // optional set of unique identifiers of parent entities.
}

type entityJSON struct {
	Type       string       `json:"type,omitempty"`
	ID         string       `json:"id,omitempty"`
	Attributes AttributeSet `json:"attributes,omitempty"`
	Parents    []string     `json:"parents,omitempty"`
}
