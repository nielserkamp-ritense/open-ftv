package models

import (
	"sync"
)

// EntitiesBuilder is the function prototype for creating a new set of entities.
type EntitiesBuilder func(in ...any) *EntitySet

// EntityIterator is the function prototype to iterate through a set of entities.
type EntityIterator func(entity *Entity)

// EntitySet represents the interface to work with a set of entities.
//
// An implementation must take care to protect against simultaneous use with concurrent go-routines.
type EntitySet struct {
	set   map[string]*Entity
	mutex sync.RWMutex
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
// E.g., only the last value with the duplicate key will be retained.
func NewEntitySet(in ...any) *EntitySet {
	out := &EntitySet{set: make(map[string]*Entity, 32)}
	for _, p := range in {
		switch t := p.(type) {
		case *Entity:
			out.set[t.UID()] = t
		case *EntitySet:
			out.mergeSet(t)
		}
	}
	return out
}

// AddEntity adds or updates an Entity in the EntitySet.
func (s *EntitySet) AddEntity(entity *Entity) (*Entity, error) {
	s.mutex.Lock()
	s.set[entity.UID()] = entity
	s.mutex.Unlock()
	return entity, nil
}

// GetEntity retrieves an Entity from the EntitySet with the given uid.
// if it exists, otherwise a nil value is returned.
func (s *EntitySet) GetEntity(uid string) *Entity {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.set[uid]
}

// RemoveEntity removes an Entity from the EntitySet.
func (s *EntitySet) RemoveEntity(uid string) {
	s.mutex.Lock()
	delete(s.set, uid)
	s.mutex.Unlock()
}

// IterateEntities iterates through all entities in the set, calling the given closure for each.
func (s *EntitySet) IterateEntities(f EntityIterator) {
	s.mutex.RLock()
	for uid := range s.set {
		f(s.set[uid])
	}
	s.mutex.RUnlock()
}

// MergeEntities merges the given EntitySet(s) into this EntitySet.
//
// Duplicate keys from an input set will overwrite the previous value.
// E.g., only the last value with the duplicate key will be retained.
func (s *EntitySet) MergeEntities(in ...*EntitySet) {
	s.mutex.Lock()
	for i := range in {
		s.mergeSet(in[i])
	}
	s.mutex.Unlock()
}

func (s *EntitySet) mergeSet(in *EntitySet) {
	in.IterateEntities(func(entity *Entity) {
		s.set[entity.UID()] = entity
	})
}
