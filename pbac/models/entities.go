package models

import (
	"fmt"
	"sync"

	"github.com/goccy/go-json"
)

// EntitiesBuilder is the function prototype for creating a new set of entities.
type EntitiesBuilder func(in ...any) EntitySet

// EntityIterator is the function prototype to iterate through a set of entities.
type EntityIterator func(entity Entity)

// Entity represents the interface to work with the details of an entity.
//
// Entity is an immutable object and is by design safe for use by concurrent go-routines.
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
		uid:     fmt.Sprintf("%s::%s", ns, id),
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

// EntitySet represents the interface to work with a set of entities.
//
// An implementation must take care to protect against simultaneous use from concurrent go-routines.
type EntitySet interface {
	AddEntity(entity Entity)          // add or replace an entity.
	GetEntity(uid string) Entity      // retrieve an entity.
	RemoveEntity(uid string)          // remove an entity.
	IterateEntities(f EntityIterator) // iterate through all entities.
	MergeEntities(in ...EntitySet)    // merge the given entity sets into this one.
}

// NewEntitySet instantiates a new standard set of attributes.
//
// It matches the EntitiesBuilder function signature.
//
// Input parameters should be of type Entity or EntitySet!
// Other types of parameters are ignored.
//
// The given entities and/or entity-sets will be copied into the returned new entity-set.
// Duplicate keys from an input set will overwrite the previous value.
// E.g. only the last value with the duplicate key will be retained.
func NewEntitySet(in ...any) EntitySet {
	out := &entities{set: make(map[string]Entity, 32)}
	for _, p := range in {
		switch t := p.(type) {
		case Entity:
			out.set[t.UID()] = t
		case EntitySet:
			out.mergeSet(t)
		}
	}
	return out
}

// AddEntity implements the EntitySety interface.
func (s *entities) AddEntity(entity Entity) {
	s.mutex.Lock()
	s.set[entity.UID()] = entity
	s.mutex.Unlock()
}

// GetEntity implements the EntitySety interface.
// if it exists, otherwise a nil value is returned.
func (s *entities) GetEntity(uid string) Entity {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.set[uid]
}

// RemoveEntity implements the EntitySety interface.
func (s *entities) RemoveEntity(uid string) {
	s.mutex.Lock()
	delete(s.set, uid)
	s.mutex.Unlock()
}

// IterateEntities implements the EntitySety interface.
func (s *entities) IterateEntities(f EntityIterator) {
	s.mutex.RLock()
	for uid := range s.set {
		f(s.set[uid])
	}
	s.mutex.RUnlock()
}

// MergeEntities implements the EntitySety interface.
//
// Duplicate keys from an input set will overwrite the previous value.
// E.g. only the last value with the duplicate key will be retained.
func (s *entities) MergeEntities(in ...EntitySet) {
	s.mutex.Lock()
	for i := range in {
		s.mergeSet(in[i])
	}
	s.mutex.Unlock()
}

func (s *entities) mergeSet(in EntitySet) {
	in.IterateEntities(func(entity Entity) {
		s.set[entity.UID()] = entity
	})
}

type entity struct {
	uid     string       // unique identifier (ns::id).
	ns      string       // name-space.
	id      string       // unique identifier within the name-space.
	attrs   AttributeSet // optional attributes.
	parents []string     // optional set of unique identifiers of parent entities.
}

type entities struct {
	set   map[string]Entity
	mutex sync.RWMutex
}

type entityJSON struct {
	Type       string       `json:"type,omitempty"`
	ID         string       `json:"id,omitempty"`
	Attributes AttributeSet `json:"attributes,omitempty"`
	Parents    []string     `json:"parents,omitempty"`
}
