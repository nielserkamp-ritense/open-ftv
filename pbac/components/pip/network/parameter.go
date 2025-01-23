package network

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/xsd"
)

// Parameter defines a parameter for a retrieval request.
//
// In should be one of:
// - "path" -> a parameter in the URI path of the request represented as :name:.
// - "query" -> a parameter in the URI query of the request.
// - "body" -> a parameter in the body of the request.
//
// Either Value + Type, or Attribute should be specified, but not both.
//
// Value represents a fixed value for the parameter.
// Type defines the type of the value.
// It should be a valid XSD datatype identifier (see http://www.w3.org/2001/XMLSchema).
// If the Type is missing, the type of value is inferred from its original input.
//
// Attribute indicates the fully qualified code of an existing attribute to be used as the value for the parameter.
type Parameter struct {
	Name        string `json:"name" yaml:"name" toml:"name"`
	In          string `json:"in" yaml:"in" toml:"in"`
	Description string `json:"description,omitempty" yaml:"description,omitempty" toml:"description,omitempty"`
	Value       any    `json:"value,omitempty" yaml:"value,omitempty" toml:"value,omitempty"`
	Type        string `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Attribute   string `json:"attribute,omitempty" yaml:"attribute,omitempty" toml:"attribute,omitempty"`
}

// ValueString returns the value of the parameter as a string.
//
// It will first attempt to retrieve the value from the Attribute code.
// If this fails, or the Attribute key is empty, Value will be returned.
//
// The returned bool indicates if the returned value is static for the parameter,
// or if it is dynamically retrieved from the current value of the attribute it points to.
func (p *Parameter) ValueString(get models.GetAttribute) (string, bool) {
	v, t, static := p.ValueAny(get)
	s, _ := xsd.ToString(v, t)
	return s, static
}

// ValueAny returns the value of the parameter.
//
// It will first attempt to retrieve the value from the Attribute code.
// If this fails, or the Attribute key is empty, Value will be returned.
//
// The returned bool indicates if the returned value is static for the parameter,
// or if it is dynamically retrieved from the current value of the attribute it points to.
func (p *Parameter) ValueAny(get models.GetAttribute) (any, string, bool) {
	switch {
	case p.Attribute != "":
		return p.AttributeValue(get)
	case p.Value != nil:
		return p.Value, p.Type, true
	default:
		return invalidValue, xsd.PrefixString, true
	}
}

// AttributeValue retrieves the attribute by the key given in the parameter and returns its value.
//
// If the attribute cannot be found, Value is returned.
//
// The returned bool indicates if the returned value is static for the parameter,
// or if it is dynamically retrieved from the current value of the attribute it points to.
func (p *Parameter) AttributeValue(get models.GetAttribute) (any, string, bool) {
	if attr := get.GetAttribute(p.Attribute); attr != nil {
		return attr.Value(), attr.Type(), false
	}

	if p.Value != nil {
		return p.Value, p.Type, true
	}
	return invalidValue, xsd.PrefixString, true
}

const invalidValue = "$invalid$"
