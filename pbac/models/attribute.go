// Package models defines constants, enumerations and models.
package models

// Attribute is a convenience type to represent a single attribute.
type Attribute struct {
	Key   string
	Value any
}

// NewAttribute instantiates a new attribute.
func NewAttribute(key string, value any) *Attribute {
	return &Attribute{Key: key, Value: value}
}
