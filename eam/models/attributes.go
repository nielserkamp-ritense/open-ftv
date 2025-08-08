package models

import (
	"slices"
	"sync"

	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
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
// E.g., only the last value with the duplicate key will be retained.
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
func (s *AttributeSet) AddAttribute(a *Attribute) (*Attribute, error) {
	s.mutex.Lock()
	s.set[a.Key()] = a
	s.mutex.Unlock()
	return a, nil
}

// AddAttributeKV adds or updates an Attribute in the AttributeSet with the specified key and value.
//
// A duplicate key will overwrite the previous value.
// E.g., only the last value with the duplicate key will be retained.
func (s *AttributeSet) AddAttributeKV(key string, value any) {
	s.AddOriginalAttribute(key, value, value, "")
}

// AddAttributeKVWithType adds or updates an Attribute in the AttributeSet with the specified key, value and type.
//
// A duplicate key will overwrite the previous value.
// E.g., only the last value with the duplicate key will be retained.
func (s *AttributeSet) AddAttributeKVWithType(key string, value any, tp string) {
	s.AddOriginalAttribute(key, value, value, tp)
}

// AddOriginalAttribute adds or updates an Attribute in the AttributeSet with a specified key, value, type and original value.
//
// A duplicate key will overwrite the previous value.
// E.g., only the last value with the duplicate key will be retained.
func (s *AttributeSet) AddOriginalAttribute(key string, value, original any, tp string) {
	if key == "" {
		return
	}

	s.mutex.Lock()
	s.set[key] = NewOriginalAttribute(key, value, original, tp)
	s.mutex.Unlock()
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
// E.g., only the last value with the duplicate key will be retained.
func (s *AttributeSet) MergeAttributes(in ...*AttributeSet) {
	s.mutex.Lock()
	for i := range in {
		in[i].IterateAttributes(func(attr *Attribute) {
			s.set[attr.Key()] = attr
		})
	}
	s.mutex.Unlock()
}

// ToOAS converts the set of attributes to a slice of OAS models.
func (s *AttributeSet) ToOAS() []attributes.Attribute {
	keys := make([]string, 0, len(s.set))
	for k := range s.set {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	out := make([]attributes.Attribute, 0, len(s.set))
	for i := range keys {
		out = append(out, *s.set[keys[i]].ToOAS())
	}
	return out
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

// AttributeSetFromOAS instantiates a new AttributeSet with the given OAS model(s).
//
// A duplicate key will overwrite the previous value.
// E.g., only the last value with the duplicate key will be retained.
func AttributeSetFromOAS(in ...any) *AttributeSet {
	s := NewAttributeSet()

	for i := range in {
		switch t := in[i].(type) {
		case attributes.Attribute:
			a := NewAttributeFromOAS(&t)
			s.set[a.Key()] = a

		case *attributes.Attribute:
			a := NewAttributeFromOAS(t)
			s.set[a.Key()] = a

		case []attributes.Attribute:
			for j := range t {
				a := NewAttributeFromOAS(&t[j])
				s.set[a.Key()] = a
			}

		case []*attributes.Attribute:
			for j := range t {
				a := NewAttributeFromOAS(t[j])
				s.set[a.Key()] = a
			}
		}
	}

	return s
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
