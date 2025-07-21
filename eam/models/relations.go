package models

import (
	"sync"
)

// RelationIterator is the function prototype to iterate through a set of relations.
type RelationIterator func(r *Relation)

// RelationSet represents a set of relations.
//
// RelationSet is safe to use across concurrent go-routines.
type RelationSet struct {
	entities *EntitySet
	set      map[string]*Relation
	mutex    sync.RWMutex
}

// NewRelationSet instantiates a new set of relations.
//
// Input parameters should be of type Relation or RelationSet!
// Other types of parameters are ignored.
//
// The given relations and/or relation-sets will be copied into the returned new relation-set.
// Duplicate keys from an input set will overwrite the previous value.
// E.g. only the last value with the duplicate key will be retained.
func NewRelationSet(entities *EntitySet, in ...any) *RelationSet {
	out := &RelationSet{entities: entities, set: make(map[string]*Relation, 32)}
	for _, p := range in {
		switch t := p.(type) {
		case *Relation:
			out.set[t.UID()] = t
		case *RelationSet:
			out.mergeSet(t)
		}
	}
	return out
}

// AddRelation adds or updates a Relation in the RelationSet.
func (s *RelationSet) AddRelation(r *Relation) {
	s.mutex.Lock()
	s.set[r.UID()] = r
	s.mutex.Unlock()
}

// AddRelationFromUID adds or updates a Relation in the RelationSet using the given UIDs.
func (s *RelationSet) AddRelationFromUID(subject, predicate, object string) {
	r := NewRelationFromUID(subject, predicate, object, s.entities)
	s.mutex.Lock()
	s.set[r.UID()] = r
	s.mutex.Unlock()
}

// GetRelation returns a Relation with the given uid from the RelationSet.
func (s *RelationSet) GetRelation(uid string) *Relation {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.set[uid]
}

// RemoveRelation removes a Relation from the RelationSet.
func (s *RelationSet) RemoveRelation(uid string) {
	s.mutex.Lock()
	delete(s.set, uid)
	s.mutex.Unlock()
}

// IterateRelations iterates through all relations in the set and calls the closure for each.
func (s *RelationSet) IterateRelations(f RelationIterator) {
	s.mutex.RLock()
	for uid := range s.set {
		f(s.set[uid])
	}
	s.mutex.RUnlock()
}

// MergeRelations merges the given RelationSet(s) into this RelationSet.
func (s *RelationSet) MergeRelations(in ...*RelationSet) {
	s.mutex.Lock()
	for i := range in {
		s.mergeSet(in[i])
	}
	s.mutex.Unlock()
}

func (s *RelationSet) mergeSet(in *RelationSet) {
	in.IterateRelations(func(r *Relation) {
		s.set[r.UID()] = r
	})
}
