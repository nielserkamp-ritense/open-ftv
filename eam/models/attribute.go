// Package schema defines generic constants, enumerations and schema.
package models

import (
	"reflect"
	"slices"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
)

// GetAttribute represents the interface for retrieving a specific attribute.
type GetAttribute interface {
	GetAttribute(key string) Attribute
}

// Attribute represents the interface for an attribute.
//
// An attribute is designed to be immutable and is thus safe for use in concurrent go-routines.
type Attribute interface {
	Key() string            // retrieve the unique key of the attribute.
	Value() any             // retrieve the derived value of the attribute.
	Original() any          // retrieve the original value of the attribute (as received from an external source).
	Type() string           // retrieve the optional type of the attribute.
	AddTags(tags ...string) // associate tags with the attribute.
	Tags() []string         // retrieve tags associated with the attribute.
	HasTag(tag string) bool // test if the attribute contains a specific tag.
	MarshalJSON() ([]byte, error)
	MarshalYAML() ([]byte, error)
}

// NewAttribute instantiates a new attribute without a specific type.
func NewAttribute(key string, value any) Attribute {
	return NewOriginalAttribute(key, value, value, "")
}

// NewAttributeWithType instantiates a new attribute with the specified type.
func NewAttributeWithType(key string, value any, tp string) Attribute {
	return NewOriginalAttribute(key, value, value, tp)
}

// NewOriginalAttribute instantiates a new attribute with the specified type and original value.
func NewOriginalAttribute(key string, value, original any, tp string) Attribute {
	return &attribute{key: key, value: value, original: original, tp: tp, tags: make(map[string]struct{})}
}

// Key implements the Attribute interface.
func (a *attribute) Key() string {
	return a.key
}

// Value implements the Attribute interface.
func (a *attribute) Value() any {
	return a.value
}

// Original implements the Attribute interface.
func (a *attribute) Original() any {
	return a.original
}

// Type implements the Attribute interface.
func (a *attribute) Type() string {
	return a.tp
}

// AddTags implements the Attribute interface.
func (a *attribute) AddTags(tags ...string) {
	for i := range tags {
		a.tags[tags[i]] = struct{}{}
	}
}

// Tags implements the Attribute interface.
func (a *attribute) Tags() []string {
	tags := make([]string, 0, len(a.tags))
	for k := range a.tags {
		tags = append(tags, k)
	}
	slices.Sort(tags)
	return tags
}

// HasTag implements the Attribute interface.
func (a *attribute) HasTag(tag string) bool {
	_, ok := a.tags[tag]
	return ok
}

// MarshalJSON implements the json.Marshaler interface.
func (a *attribute) MarshalJSON() ([]byte, error) {
	m := marshalAttr{
		Key:   a.key,
		Value: a.value,
		Type:  a.tp,
		Tags:  a.Tags(),
	}

	if a.original != a.value {
		m.Original = a.original
	}

	return json.Marshal(m)
}

// MarshalYAML implements the yaml.Marshaler interface.
func (a *attribute) MarshalYAML() ([]byte, error) {
	m := marshalAttr{
		Key:      a.key,
		Value:    a.value,
		Original: a.original,
		Type:     a.tp,
		Tags:     a.Tags(),
	}
	return yaml.Marshal(m)
}

// AttributeEqual returns true if the given attributes are equal.
func AttributeEqual(a1, a2 Attribute) bool {
	return a1.Key() == a2.Key() &&
		a1.Type() == a2.Type() &&
		reflect.DeepEqual(a1.Value(), a2.Value()) &&
		reflect.DeepEqual(a1.Original(), a2.Original()) &&
		reflect.DeepEqual(a1.Tags(), a2.Tags())
}

type attribute struct {
	key      string
	value    any
	original any
	tp       string
	tags     map[string]struct{}
}

type marshalAttr struct {
	Key      string   `json:"key"                yaml:"key"`
	Value    any      `json:"value"              yaml:"value"`
	Original any      `json:"original,omitempty" yaml:"original,omitempty"`
	Type     string   `json:"type,omitempty"     yaml:"type,omitempty"`
	Tags     []string `json:"tags,omitempty"     yaml:"tags,omitempty"`
}
