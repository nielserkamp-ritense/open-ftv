// Package models defines generic constants, enumerations and models.
package models

import "github.com/goccy/go-json"

// Attribute is a convenience type to represent a single attribute.
type Attribute interface {
	Key() string
	Value() any
	Type() string
}

// NewAttribute instantiates a new attribute without a specific type.
func NewAttribute(key string, value any) Attribute {
	return &attribute{key: key, value: value}
}

// NewAttributeWithType instantiates a new attribute with the specified type.
func NewAttributeWithType(key string, value any, t string) Attribute {
	return &attribute{key: key, value: value, tp: t}
}

// Key implements the Attribute interface.
func (a *attribute) Key() string {
	return a.key
}

// Value implements the Attribute interface.
func (a *attribute) Value() any {
	return a.value
}

// Type implements the Attribute interface.
func (a *attribute) Type() string {
	return a.tp
}

// MarshalJSON implements the json.Marshaller interface.
func (a *attribute) MarshalJSON() ([]byte, error) {
	return json.Marshal(attributeJSON{
		Key:   a.key,
		Value: a.value,
		Type:  a.tp,
	})
}

type attribute struct {
	key   string
	value any
	tp    string
}

type attributeJSON struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
	Type  string `json:"type,omitempty"`
}
