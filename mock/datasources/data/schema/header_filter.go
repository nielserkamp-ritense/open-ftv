package schema

import (
	"fmt"
	"regexp"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
)

// HeaderFilter limits the results of an endpoint to the rows where a field equals a value taken from a request header.
//
// The header value must match the pattern; the first capture group of the pattern is the value to filter on.
type HeaderFilter struct {
	Header  string // the request header to read.
	Field   string // the field to filter on; qualify it with the table id ("t1.f1") for a joined table.
	Pattern string // the regular expression the header value must match.
	// hidden fields
	rx *regexp.Regexp
}

// NewHeaderFilter instantiates a header filter, failing if the pattern doesn't compile or has no capture group.
func NewHeaderFilter(header, field, pattern string) (*HeaderFilter, error) {
	rx, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("header filter %s: %w", header, err)
	}

	if rx.NumSubexp() < 1 {
		return nil, fmt.Errorf("header filter %s: pattern %q has no capture group", header, pattern)
	}

	return &HeaderFilter{Header: header, Field: field, Pattern: pattern, rx: rx}, nil
}

// Value returns the value to filter on, taken from the given header value.
func (h *HeaderFilter) Value(header string) (string, error) {
	if header == "" {
		return "", fmt.Errorf("missing header %s", h.Header)
	}

	m := h.rx.FindStringSubmatch(header)
	if m == nil {
		return "", fmt.Errorf("invalid header %s: %q does not match %q", h.Header, header, h.Pattern)
	}

	return m[1], nil
}

// MarshalJSON implements the JSON Marshaler interface.
func (h *HeaderFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(&encodeHeaderFilter{Header: h.Header, Field: h.Field, Pattern: h.Pattern})
}

// UnmarshalJSON implements the JSON Unmarshaler interface.
func (h *HeaderFilter) UnmarshalJSON(b []byte) error {
	var h2 encodeHeaderFilter
	if err := json.Unmarshal(b, &h2); err != nil {
		return err
	}

	return h.set(h2)
}

// MarshalYAML implements the YAML Marshaler interface.
func (h *HeaderFilter) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(&encodeHeaderFilter{Header: h.Header, Field: h.Field, Pattern: h.Pattern})
}

// UnmarshalYAML implements the YAML Unmarshaler interface.
func (h *HeaderFilter) UnmarshalYAML(b []byte) error {
	var h2 encodeHeaderFilter
	if err := yaml.Unmarshal(b, &h2); err != nil {
		return err
	}

	return h.set(h2)
}

func (h *HeaderFilter) set(h2 encodeHeaderFilter) error {
	h3, err := NewHeaderFilter(h2.Header, h2.Field, h2.Pattern)
	if err != nil {
		return err
	}

	*h = *h3

	return nil
}

type encodeHeaderFilter struct {
	Header  string `json:"header" yaml:"header"`
	Field   string `json:"field" yaml:"field"`
	Pattern string `json:"pattern" yaml:"pattern"`
}
