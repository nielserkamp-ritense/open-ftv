// Package models defines generic constants, enumerations and models.
package models

import (
	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
)

// Attribute represents the interface for an attribute.
//
// Attribute is an immutable object and is by design safe for use in concurrent go-routines.
type Attribute interface {
	Key() string   // retrieve the unique key of the attribute.
	Value() any    // retrieve the derived value of the attribute.
	Original() any // retrieve the original value of the attribute (as received from external source).
	Type() string  // retrieve the optional type of the attribute.
	MarshalJSON() ([]byte, error)
	MarshalYAML() ([]byte, error)
}

// NewAttribute instantiates a new attribute without a specific type.
func NewAttribute(key string, value any) Attribute {
	return &attribute{key: key, value: value}
}

// NewAttributeWithType instantiates a new attribute with the specified type.
func NewAttributeWithType(key string, value any, tp string) Attribute {
	return &attribute{key: key, value: value, tp: tp}
}

// NewOriginalAttribute instantiates a new attribute with the specified type and original value.
func NewOriginalAttribute(key string, value, original any, tp string) Attribute {
	return &attribute{key: key, value: value, original: original, tp: tp}
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

// MarshalJSON implements the json.Marshaller interface.
func (a *attribute) MarshalJSON() ([]byte, error) {
	return json.Marshal(marshalAttr{
		Key:      a.key,
		Value:    a.value,
		Original: a.original,
		Type:     a.tp,
	})
}

// MarshalYAML implements the yaml.Marshaller interface.
func (a *attribute) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(marshalAttr{
		Key:      a.key,
		Value:    a.value,
		Original: a.original,
		Type:     a.tp,
	})
}

type attribute struct {
	key      string
	value    any
	original any
	tp       string
}

type marshalAttr struct {
	Key      string `json:"key" yaml:"key"`
	Value    any    `json:"value" yaml:"value"`
	Original any    `json:"original,omitempty" yaml:"original,omitempty"`
	Type     string `json:"type,omitempty" yaml:"type,omitempty"`
}
