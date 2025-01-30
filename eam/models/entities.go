package models

import (
	"sync"
)

// EntitiesBuilder is the function prototype for creating a new set of entities.
type EntitiesBuilder func(in ...any) EntitySet

// EntityIterator is the function prototype to iterate through a set of entities.
type EntityIterator func(entity Entity)

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

type entities struct {
	set   map[string]Entity
	mutex sync.RWMutex
}
