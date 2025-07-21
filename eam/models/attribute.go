// Package schema defines generic constants, enumerations and schema.
package models

import (
	"reflect"
	"slices"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
)

// Attribute contains the details of an attribute.
//
// An attribute is designed to be immutable and is thus safe for use in concurrent go-routines.
type Attribute struct {
	key      string
	value    any
	original any
	tp       string
	tags     map[string]struct{}
}

// NewAttribute instantiates a new attribute without a specific type.
func NewAttribute(key string, value any) *Attribute {
	return NewOriginalAttribute(key, value, value, "")
}

// NewAttributeWithType instantiates a new attribute with a specific type.
func NewAttributeWithType(key string, value any, tp string) *Attribute {
	return NewOriginalAttribute(key, value, value, tp)
}

// NewOriginalAttribute instantiates a new attribute with a specific type and original value.
func NewOriginalAttribute(key string, value, original any, tp string) *Attribute {
	return &Attribute{key: key, value: value, original: original, tp: tp, tags: make(map[string]struct{})}
}

// Key returns the key of the Attribute.
func (a *Attribute) Key() string {
	return a.key
}

// Value returns the value of the Attribute.
func (a *Attribute) Value() any {
	return a.value
}

// Original returns the original value of the Attribute.
func (a *Attribute) Original() any {
	return a.original
}

// Type returns the type of Attribute.
func (a *Attribute) Type() string {
	return a.tp
}

// AddTags associates the given tags with the Attribute.
func (a *Attribute) AddTags(tags ...string) {
	for i := range tags {
		a.tags[tags[i]] = struct{}{}
	}
}

// Tags returns the tags for the Attribute.
func (a *Attribute) Tags() []string {
	tags := make([]string, 0, len(a.tags))
	for k := range a.tags {
		tags = append(tags, k)
	}
	slices.Sort(tags)
	return tags
}

// HasTag reutns true if the Attribute is associated with the given tag.
func (a *Attribute) HasTag(tag string) bool {
	_, ok := a.tags[tag]
	return ok
}

// MarshalJSON implements the json.Marshaler interface.
func (a *Attribute) MarshalJSON() ([]byte, error) {
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
func (a *Attribute) MarshalYAML() ([]byte, error) {
	m := marshalAttr{
		Key:      a.key,
		Value:    a.value,
		Original: a.original,
		Type:     a.tp,
		Tags:     a.Tags(),
	}
	return yaml.Marshal(m)
}

// Equals returns true if this Attribute equals the other Attribute.
func (a *Attribute) Equals(other *Attribute) bool {
	return a.key == other.key &&
		a.tp == other.tp &&
		reflect.DeepEqual(a.value, other.value) &&
		reflect.DeepEqual(a.original, other.original) &&
		reflect.DeepEqual(a.tags, other.tags)
}

type marshalAttr struct {
	Key      string   `json:"key"                yaml:"key"`
	Value    any      `json:"value"              yaml:"value"`
	Original any      `json:"original,omitempty" yaml:"original,omitempty"`
	Type     string   `json:"type,omitempty"     yaml:"type,omitempty"`
	Tags     []string `json:"tags,omitempty"     yaml:"tags,omitempty"`
}
