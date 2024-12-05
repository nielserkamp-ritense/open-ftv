package models

import (
	"fmt"
	"sync"
)

// Relation represents the details of a relationship between a subject and an object.
//
// Relation is an immutable object and is by design safe for use by concurrent go-routines.
type Relation interface {
	UID() string       // return the unique identifier of this relation.
	Subject() Entity   // returns the subject of the relation.
	Predicate() Entity // returns the predicate of the relation.
	Object() Entity    // returns the object of the relation.
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
		uid:       fmt.Sprintf("%s|%s|%s", subject.UID(), predicate.UID(), object.UID()),
		subject:   subject,
		predicate: predicate,
		object:    object,
	}
}

// UID implements the Relation interface.
func (r *relation) UID() string { return r.uid }

// Subject implements the Relation interface.
func (r *relation) Subject() Entity { return r.subject }

// Predicate implements the Relation interface.
func (r *relation) Predicate() Entity { return r.predicate }

// Object implements the Relation interface.
func (r *relation) Object() Entity { return r.object }

// RelationIterator is the function prototype to iterate through a set of relations.
type RelationIterator func(r Relation)

// RelationSet represents a set of relations.
//
// RelationSet is safe to use across concurrent go-routines.
type RelationSet interface {
	AddRelation(r Relation)                               // add or replace a relation.
	AddRelationFromUID(subject, predicate, object string) // add or replace a relation using the given keys.
	GetRelation(uid string) any                           // retrieve a relation.
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
func (s *relations) GetRelation(uid string) any {
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

type relation struct {
	uid       string
	subject   Entity
	predicate Entity
	object    Entity
}

type relations struct {
	entities EntitySet
	set      map[string]Relation
	mutex    sync.RWMutex
}
