package filtering

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"

// Filterer represents the interface for testing a filter or set of filters on a set of key/value pairs (a row of data).
type Filterer interface {
	String() string
	Prepare(ds *schema.Datasource, joins []*schema.Join) error
	MatchOnPrimaryData(data map[string]any) bool
	MatchOnJoinData(join *schema.Join, data map[string]any) bool
}
