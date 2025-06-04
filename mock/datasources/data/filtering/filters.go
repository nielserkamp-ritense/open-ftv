package filtering

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

// Filters is a convenience type for a list of filters.
type Filters []*Filter

// Prepare implements the Filterer interface.
func (f Filters) Prepare(ds *schema.Datasource, joins []*schema.Join) error {
	for _, filter := range f {
		if err := filter.Prepare(ds, joins); err != nil {
			return err
		}
	}
	return nil
}
