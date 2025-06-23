package filtering

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

// AllOfFilter represents an "and" filter, which means it matches if all the underlying filters are matched.
type AllOfFilter struct {
	AllOf Filters `json:"allOf" yaml:"allOf"`
}

// String implements the Stringer interface.
func (a *AllOfFilter) String() string {
	return fmt.Sprintf("{allOf:%v}", a.AllOf)
}

// Prepare implements the Filterer interface.
func (a *AllOfFilter) Prepare(ds *schema.Datasource, joins []*schema.Join) error {
	return a.AllOf.Prepare(ds, joins)
}

// MatchOnPrimaryData implements the Filterer interface.
func (f *AllOfFilter) MatchOnPrimaryData(data map[string]any) bool {
	for _, filter := range f.AllOf {
		if ok := filter.MatchOnPrimaryData(data); !ok {
			return false
		}
	}
	return true
}

// MatchOnJoinData implements the Filterer interface.
func (f *AllOfFilter) MatchOnJoinData(join *schema.Join, data map[string]any) bool {
	for _, filter := range f.AllOf {
		if ok := filter.MatchOnJoinData(join, data); !ok {
			return false
		}
	}
	return true
}
