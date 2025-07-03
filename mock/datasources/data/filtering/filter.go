package filtering

import (
	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
)

// Filter represents either a single field/value filter or a set of filters with AllOf or AnyOf functionality.
type Filter struct {
	FieldValueFilter
	AllOfFilter
	AnyOfFilter
}

// FilterFromAny attempts to convert the given input to a filter.
//
// This function will attempt to decode the input in the following sequence:
// - as JSON format
// - as YAML format
// - as text format (work-in-progress)
func FilterFromAny(primary, in string) *Filter {
	out := new(Filter)

	// try JSON decoding.
	if err := out.UnmarshalJSON([]byte(in)); err == nil {
		return out
	}

	// try YAML decoding.
	if err := out.UnmarshalYAML([]byte(in)); err == nil {
		return out
	}

	return FilterFromString(primary, in)
}

// String implements the Stringer interface.
func (f *Filter) String() string {
	switch {
	case f.AllOf != nil:
		return f.AllOfFilter.String()
	case f.AnyOf != nil:
		return f.AnyOfFilter.String()
	default:
		return f.FieldValueFilter.String()
	}
}

// Prepare implements the Filterer interface.
func (f *Filter) Prepare(ds *schema.Datasource, joins []*schema.Join) error {
	switch {
	case f.AllOf != nil:
		return f.AllOfFilter.Prepare(ds, joins)
	case f.AnyOf != nil:
		return f.AnyOfFilter.Prepare(ds, joins)
	default:
		return f.FieldValueFilter.Prepare(ds, joins)
	}
}

// MatchOnPrimaryData implements the Filterer interface.
func (f *Filter) MatchOnPrimaryData(data map[string]any) bool {
	switch {
	case f.AllOf != nil:
		return f.AllOfFilter.MatchOnPrimaryData(data)
	case f.AnyOf != nil:
		return f.AnyOfFilter.MatchOnPrimaryData(data)
	default:
		return f.FieldValueFilter.MatchOnPrimaryData(data)
	}
}

// MatchOnJoinData implements the Filterer interface.
func (f *Filter) MatchOnJoinData(join *schema.Join, data map[string]any) bool {
	switch {
	case f.AllOf != nil:
		return f.AllOfFilter.MatchOnJoinData(join, data)
	case f.AnyOf != nil:
		return f.AnyOfFilter.MatchOnJoinData(join, data)
	default:
		return f.FieldValueFilter.MatchOnJoinData(join, data)
	}
}

// UnmarshalJSON implements the JSON unmarshaler interface.
func (f *Filter) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &f.AllOfFilter); err == nil && f.AllOf != nil {
		return nil
	}
	if err := json.Unmarshal(data, &f.AnyOfFilter); err == nil && f.AnyOf != nil {
		return nil
	}
	return json.Unmarshal(data, &f.FieldValueFilter)
}

// UnmarshalYAML implements the YAML unmarshaler interface.
func (f *Filter) UnmarshalYAML(data []byte) error {
	if err := yaml.Unmarshal(data, &f.AllOfFilter); err == nil && f.AllOf != nil {
		return nil
	}
	if err := yaml.Unmarshal(data, &f.AnyOfFilter); err == nil && f.AnyOf != nil {
		return nil
	}
	return yaml.Unmarshal(data, &f.FieldValueFilter)
}
