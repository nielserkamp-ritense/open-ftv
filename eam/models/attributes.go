package models

import (
	"slices"
	"strings"
	"sync"

	"github.com/goccy/go-json"
)

// AttributesBuilder is the function prototype for creating a new set of attributes.
type AttributesBuilder func(in ...any) AttributeSet

// AttributeIterator is the function prototype to iterate through a set of attributes.
type AttributeIterator func(attr Attribute)

// AttributeSet represents the interface to work with a set of attributes.
//
// An implementation must take care to protect against simultaneous use from concurrent go-routines.
type AttributeSet interface {
	AddAttribute(key string, value any)                              // add or replace an attribute without a specific type.
	AddAttributeWithType(key string, value any, tp string)           // add or replace an attribute with the specified type.
	AddOriginalAttribute(key string, value, original any, tp string) // add or replace an attribute with the specified type and original value.
	GetAttribute(key string) Attribute                               // retrieve an attribute.
	GetAttributeValue(key string) any                                // retrieve an attribute value.
	RemoveAttribute(key string)                                      // remove an attribute.
	IterateAttributes(f AttributeIterator)                           // iterate through all attributes.
	MergeAttributes(in ...AttributeSet)                              // merge the given attribute sets into this one.
	MarshalJSON() ([]byte, error)
}

// NewAttributeSet instantiates a new set of attributes.
//
// It matches the AttributesBuilder function signature.
//
// The given attribute sets will be copied into the returned new attribute set.
// Duplicate keys from an input set will overwrite the previous value.
// E.g. only the last value with the duplicate key will be retained.
func NewAttributeSet(in ...any) AttributeSet {
	out := &attributes{set: make(map[string]Attribute, 32)}

	for _, p := range in {
		switch t := p.(type) {
		case Attribute:
			out.set[t.Key()] = t
		case AttributeSet:
			t.IterateAttributes(func(attr Attribute) {
				out.set[attr.Key()] = attr
			})
		case []Attribute:
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

// AddAttribute implements the AttributeSet interface.
//
// A duplicate key will overwrite the previous value.
// E.g. only the last value with the duplicate key will be retained.
func (a *attributes) AddAttribute(key string, value any) {
	a.AddOriginalAttribute(key, value, value, "")
}

// AddAttributeWithType implements the AttributeSet interface.
//
// A duplicate key will overwrite the previous value.
// E.g. only the last value with the duplicate key will be retained.
func (a *attributes) AddAttributeWithType(key string, value any, tp string) {
	a.AddOriginalAttribute(key, value, value, tp)
}

// AddOriginalAttribute implements the AttributeSet interface.
//
// A duplicate key will overwrite the previous value.
// E.g. only the last value with the duplicate key will be retained.
//
// If the key contains dots, it is considered to be a multi-level attribute key.
// Any key at a level that does not exist in the attribute tree will be created.
// If an existing key level does not have an object as value, its value will be replaced with an object.
func (a *attributes) AddOriginalAttribute(key string, value, original any, tp string) {
	if key == "" {
		return
	}

	keys := strings.Split(key, ".")

	a.mutex.Lock()
	a.addOriginalAttribute(keys, value, original, tp)
	a.mutex.Unlock()
}

func (a *attributes) addOriginalAttribute(keys []string, value, original any, tp string) {
	key := keys[0]
	if len(keys) == 1 {
		a.set[key] = NewOriginalAttribute(key, value, original, tp)
		return
	}

	upsert := func() *attributes {
		set := NewAttributeSet()
		a.set[key] = NewAttribute(key, set)
		return set.(*attributes)
	}

	var set *attributes
	if attr := a.getAttribute(key); attr == nil {
		set = upsert()
	} else {
		var ok bool
		if set, ok = attr.Value().(*attributes); !ok {
			set = upsert()
		}
	}

	set.addOriginalAttribute(keys[1:], value, original, tp)
}

// GetAttribute implements the AttributeSet interface.
func (a *attributes) GetAttribute(key string) Attribute {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.getAttribute(key)
}

func (a *attributes) getAttribute(key string) Attribute {
	return a.set[key]
}

// GetAttributeValue implements the AttributeSet interface.
func (a *attributes) GetAttributeValue(key string) any {
	if attr := a.GetAttribute(key); attr != nil {
		return attr.Value()
	}
	return nil
}

// RemoveAttribute implements the AttributeSet interface.
func (a *attributes) RemoveAttribute(key string) {
	a.mutex.Lock()
	delete(a.set, key)
	a.mutex.Unlock()
}

// IterateAttributes implements the AttributeSet interface.
//
// The supplied callback function should be as short-lived as possible,
// as this function locks the attribute set against any updates.
func (a *attributes) IterateAttributes(f AttributeIterator) {
	a.mutex.RLock()
	for k := range a.set {
		f(a.set[k])
	}
	a.mutex.RUnlock()
}

// MergeAttributes implements the AttributeSet interface.
//
// Duplicate keys from an input set will overwrite the previous value.
// E.g. only the last value with the duplicate key will be retained.
func (a *attributes) MergeAttributes(in ...AttributeSet) {
	a.mutex.Lock()
	for i := range in {
		in[i].IterateAttributes(func(attr Attribute) {
			a.set[attr.Key()] = attr
		})
	}
	a.mutex.Unlock()
}

// MarshalJSON implements the json.Marshaller interface.
func (a *attributes) MarshalJSON() ([]byte, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	keys := make([]string, 0, len(a.set))
	for k := range a.set {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	list := make([]Attribute, 0, len(a.set))
	for i := range keys {
		list = append(list, a.set[keys[i]])
	}

	return json.Marshal(list)
}

// MapFromAttributes returns a standard map from the given attribute set.
func MapFromAttributes(in AttributeSet) map[string]any {
	if in == nil {
		return nil
	}

	out := make(map[string]any)
	in.IterateAttributes(func(attr Attribute) {
		out[attr.Key()] = attr.Value()
	})

	return out
}

// AttributesEqual returns true if the given sets of attributes match exactly.
func AttributesEqual(s1, s2 AttributeSet) bool {
	if s1 == nil && s2 == nil {
		return true
	}

	as1, ok1 := s1.(*attributes)
	as2, ok2 := s2.(*attributes)

	if !ok1 || !ok2 {
		return false
	}

	as1.mutex.RLock()
	defer as1.mutex.RUnlock()

	as2.mutex.RLock()
	defer as2.mutex.RUnlock()

	for _, a1 := range as1.set {
		a2 := as2.getAttribute(a1.Key())
		if a2 == nil || !AttributeEqual(a1, a2) {
			return false
		}
	}

	for _, a1 := range as2.set {
		a2 := as1.getAttribute(a1.Key())
		if a2 == nil || !AttributeEqual(a1, a2) {
			return false
		}
	}

	return true
}

func (a *attributes) addMap(m map[string]any) {
	for k := range m {
		a.set[k] = NewAttribute(k, m[k])
	}
}

type attributes struct {
	set   map[string]Attribute
	mutex sync.RWMutex
}
