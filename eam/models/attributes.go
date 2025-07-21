package models

import (
	"slices"
	"strings"
	"sync"

	"github.com/goccy/go-json"
)

// AttributesBuilder is the function prototype for creating a new set of attributes.
type AttributesBuilder func(in ...any) *AttributeSet

// AttributeIterator is the function prototype to iterate through a set of attributes.
type AttributeIterator func(attr *Attribute)

// AttributeSet represents the interface to work with a set of attributes.
//
// An implementation must take care to protect against simultaneous use from concurrent go-routines.
type AttributeSet struct {
	set   map[string]*Attribute
	mutex sync.RWMutex
}

// NewAttributeSet instantiates a new set of attributes.
//
// It matches the AttributesBuilder function signature.
//
// The given attribute sets will be copied into the returned new attribute set.
// Duplicate keys from an input set will overwrite the previous value.
// E.g. only the last value with the duplicate key will be retained.
func NewAttributeSet(in ...any) *AttributeSet {
	out := &AttributeSet{set: make(map[string]*Attribute, 32)}

	for _, p := range in {
		switch t := p.(type) {
		case *Attribute:
			out.set[t.Key()] = t
		case *AttributeSet:
			if t != nil {
				t.IterateAttributes(func(attr *Attribute) {
					out.set[attr.Key()] = attr
				})
			}
		case []*Attribute:
			for i := range t {
				attr := t[i]
				out.set[attr.Key()] = attr
			}
		case map[string]any:
			out.addMap(t)
		case *map[string]any:
			if t != nil {
				out.addMap(*t)
			}
		}
	}
	return out
}

// AddAttribute adds or updates an Attribute in the AttributeSet.
//
// A duplicate key will overwrite the previous value.
// E.g., only the last value with the duplicate key will be retained.
func (s *AttributeSet) AddAttribute(key string, value any) {
	s.AddOriginalAttribute(key, value, value, "")
}

// AddAttributeWithType adds or updates an Attribute in the AttributeSet with a specific type.
//
// A duplicate key will overwrite the previous value.
// E.g., only the last value with the duplicate key will be retained.
func (s *AttributeSet) AddAttributeWithType(key string, value any, tp string) {
	s.AddOriginalAttribute(key, value, value, tp)
}

// AddOriginalAttribute adds or updates an Attribute in the AttributeSet with a specific type and original value.
//
// A duplicate key will overwrite the previous value.
// E.g., only the last value with the duplicate key will be retained.
//
// If the key contains dots, it is considered to be a multi-level attribute key.
// Any key at a level that does not exist in the attribute tree will be created.
// If an existing key level does not have an object as value, its value will be replaced with an object.
func (s *AttributeSet) AddOriginalAttribute(key string, value, original any, tp string) {
	if key == "" {
		return
	}

	keys := strings.Split(key, ".")

	s.mutex.Lock()
	s.addOriginalAttribute(keys, value, original, tp)
	s.mutex.Unlock()
}

func (s *AttributeSet) addOriginalAttribute(keys []string, value, original any, tp string) {
	key := keys[0]
	if len(keys) == 1 {
		s.set[key] = NewOriginalAttribute(key, value, original, tp)
		return
	}

	upsert := func() *AttributeSet {
		set := NewAttributeSet()
		s.set[key] = NewAttribute(key, set)
		return set
	}

	var set *AttributeSet
	if attr := s.getAttribute(key); attr == nil {
		set = upsert()
	} else {
		var ok bool
		if set, ok = attr.Value().(*AttributeSet); !ok {
			set = upsert()
		}
	}

	set.addOriginalAttribute(keys[1:], value, original, tp)
}

// GetAttribute retrieves an Attribute from the AttributeSet with the given uid.
func (s *AttributeSet) GetAttribute(key string) *Attribute {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.getAttribute(key)
}

func (s *AttributeSet) getAttribute(key string) *Attribute {
	return s.set[key]
}

// GetAttributeValue retrieves an Attribute value from the AttributeSet with the given uid.
func (s *AttributeSet) GetAttributeValue(key string) any {
	if attr := s.GetAttribute(key); attr != nil {
		return attr.Value()
	}
	return nil
}

// RemoveAttribute removes an Attribute from the AttributeSet.
func (s *AttributeSet) RemoveAttribute(key string) {
	s.mutex.Lock()
	delete(s.set, key)
	s.mutex.Unlock()
}

// IterateAttributes iterates through all attributes in the set and calls the given closure for each.
//
// The supplied callback function should be as short-lived as possible,
// as this function locks the attribute set against any updates.
func (s *AttributeSet) IterateAttributes(f AttributeIterator) {
	s.mutex.RLock()
	for k := range s.set {
		f(s.set[k])
	}
	s.mutex.RUnlock()
}

// MergeAttributes merges the given AttributeSet(s) into this AttributeSet.
//
// Duplicate keys from an input set will overwrite the previous value.
// E.g. only the last value with the duplicate key will be retained.
func (s *AttributeSet) MergeAttributes(in ...*AttributeSet) {
	s.mutex.Lock()
	for i := range in {
		in[i].IterateAttributes(func(attr *Attribute) {
			s.set[attr.Key()] = attr
		})
	}
	s.mutex.Unlock()
}

// MarshalJSON implements the json.Marshaler interface.
func (s *AttributeSet) MarshalJSON() ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	keys := make([]string, 0, len(s.set))
	for k := range s.set {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	list := make([]*Attribute, 0, len(s.set))
	for i := range keys {
		list = append(list, s.set[keys[i]])
	}

	return json.Marshal(list)
}

// MapFromAttributes returns a standard map from the given attribute set.
func MapFromAttributes(in *AttributeSet) map[string]any {
	if in == nil {
		return nil
	}

	out := make(map[string]any)
	in.IterateAttributes(func(attr *Attribute) {
		out[attr.Key()] = attr.Value()
	})

	return out
}

// Equals returns true if this AttributeSet matches exactly the other AttributeSet.
func (s *AttributeSet) Equals(other *AttributeSet) bool {
	if s == nil && other == nil {
		return true
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	other.mutex.RLock()
	defer other.mutex.RUnlock()

	for _, a1 := range s.set {
		a2 := other.getAttribute(a1.Key())
		if a2 == nil || !a1.Equals(a2) {
			return false
		}
	}

	for _, a1 := range other.set {
		a2 := s.getAttribute(a1.Key())
		if a2 == nil || !a1.Equals(a2) {
			return false
		}
	}

	return true
}

func (s *AttributeSet) addMap(m map[string]any) {
	for k := range m {
		s.set[k] = NewAttribute(k, m[k])
	}
}
