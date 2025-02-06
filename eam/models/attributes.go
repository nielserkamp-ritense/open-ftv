package models

import (
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
	a.mutex.Lock()
	a.set[key] = NewAttribute(key, value)
	a.mutex.Unlock()
}

// AddAttributeWithType implements the AttributeSet interface.
//
// A duplicate key will overwrite the previous value.
// E.g. only the last value with the duplicate key will be retained.
func (a *attributes) AddAttributeWithType(key string, value any, tp string) {
	a.mutex.Lock()
	a.set[key] = NewAttributeWithType(key, value, tp)
	a.mutex.Unlock()
}

// AddOriginalAttribute implements the AttributeSet interface.
//
// A duplicate key will overwrite the previous value.
// E.g. only the last value with the duplicate key will be retained.
func (a *attributes) AddOriginalAttribute(key string, value, original any, tp string) {
	a.mutex.Lock()
	a.set[key] = NewOriginalAttribute(key, value, original, tp)
	a.mutex.Unlock()
}

// GetAttribute implements the AttributeSet interface.
func (a *attributes) GetAttribute(key string) Attribute {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.set[key]
}

// GetAttributeValue implements the AttributeSet interface.
func (a *attributes) GetAttributeValue(key string) any {
	a.mutex.RLock()
	attr := a.set[key]
	a.mutex.RUnlock()

	if attr != nil {
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
	out := make(map[string]any, len(a.set))
	for k := range a.set {
		out[k] = a.set[k].Value()
	}
	return json.Marshal(out)
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

func (a *attributes) addMap(m map[string]any) {
	for k := range m {
		a.set[k] = NewAttribute(k, m[k])
	}
}

type attributes struct {
	set   map[string]Attribute
	mutex sync.RWMutex
}
