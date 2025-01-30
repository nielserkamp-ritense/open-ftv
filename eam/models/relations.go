package models

import (
	"sync"
)

// RelationIterator is the function prototype to iterate through a set of relations.
type RelationIterator func(r Relation)

// RelationSet represents a set of relations.
//
// RelationSet is safe to use across concurrent go-routines.
type RelationSet interface {
	AddRelation(r Relation)                               // add or replace a relation.
	AddRelationFromUID(subject, predicate, object string) // add or replace a relation using the given keys.
	GetRelation(uid string) Relation                      // retrieve a relation.
	RemoveRelation(uid string)                            // remove a relation.
	IterateRelations(f RelationIterator)                  // iterate through all relations.
	MergeRelations(in ...RelationSet)                     // merge the given relation sets into this one.
}

// NewRelationSet instantiates a new set of relations.
//
// Input parameters should be of type Relation or RelationSet!
// Other types of parameters are ignored.
//
// The given relations and/or relation-sets will be copied into the returned new relation-set.
// Duplicate keys from an input set will overwrite the previous value.
// E.g. only the last value with the duplicate key will be retained.
func NewRelationSet(entities EntitySet, in ...any) RelationSet {
	out := &relations{entities: entities, set: make(map[string]Relation, 32)}
	for _, p := range in {
		switch t := p.(type) {
		case Relation:
			out.set[t.UID()] = t
		case RelationSet:
			out.mergeSet(t)
		}
	}
	return out
}

// AddRelation implements the RelationSet interface.
func (s *relations) AddRelation(r Relation) {
	s.mutex.Lock()
	s.set[r.UID()] = r
	s.mutex.Unlock()
}

// AddRelationFromUID implements the RelationSet interface.
func (s *relations) AddRelationFromUID(subject, predicate, object string) {
	r := NewRelationFromUID(subject, predicate, object, s.entities)
	s.mutex.Lock()
	s.set[r.UID()] = r
	s.mutex.Unlock()
}

// GetRelation implements the RelationSet interface.
func (s *relations) GetRelation(uid string) Relation {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.set[uid]
}

// RemoveRelation implements the RelationSet interface.
func (s *relations) RemoveRelation(uid string) {
	s.mutex.Lock()
	delete(s.set, uid)
	s.mutex.Unlock()
}

// IterateRelations implements the RelationSet interface.
func (s *relations) IterateRelations(f RelationIterator) {
	s.mutex.RLock()
	for uid := range s.set {
		f(s.set[uid])
	}
	s.mutex.RUnlock()
}

// MergeRelations implements the RelationSet interface.
func (s *relations) MergeRelations(in ...RelationSet) {
	s.mutex.Lock()
	for i := range in {
		s.mergeSet(in[i])
	}
	s.mutex.Unlock()
}

func (s *relations) mergeSet(in RelationSet) {
	in.IterateRelations(func(r Relation) {
		s.set[r.UID()] = r
	})
}

type relations struct {
	entities EntitySet
	set      map[string]Relation
	mutex    sync.RWMutex
}
