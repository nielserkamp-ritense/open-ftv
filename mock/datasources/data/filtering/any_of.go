package filtering

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

// AnyOfFilter represents an "or" filter, which means it matches if one of the underlying filters is matched.
type AnyOfFilter struct {
	AnyOf Filters `json:"anyOf" yaml:"anyOf"`
}

// String implements the Stringer interface.
func (a *AnyOfFilter) String() string {
	return fmt.Sprintf("AnyOf: %v", a.AnyOf)
}

// Prepare implements the Filterer interface.
func (a *AnyOfFilter) Prepare(ds *schema.Datasource, joins []*schema.Join) error {
	return a.AnyOf.Prepare(ds, joins)
}

// MatchOnPrimaryData implements the Filterer interface.
func (f *AnyOfFilter) MatchOnPrimaryData(data map[string]any) bool {
	for _, filter := range f.AnyOf {
		if ok := filter.MatchOnPrimaryData(data); ok {
			return true
		}
	}
	return false
}

// MatchOnJoinData implements the Filterer interface.
func (f *AnyOfFilter) MatchOnJoinData(join *schema.Join, data map[string]any) bool {
	for _, filter := range f.AnyOf {
		if ok := filter.MatchOnJoinData(join, data); ok {
			return true
		}
	}
	return false
}
