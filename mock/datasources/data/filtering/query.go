package filtering

import (
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/maps"
)

// FilterFromQuery returns a filter based on a previously decoded HTTP query string.
func FilterFromQuery(ds *schema.Datasource, query map[string]string) Filterer {
	out := &Filter{AllOfFilter: AllOfFilter{AllOf: make(Filters, 0)}}

	makeFilter := func(table, key, value string) *Filter {
		filter := Compare(key, enums.IsEqual, value)
		if ds != nil && len(table) > 0 {
			if ds.Table(table) != nil {
				filter = filter.OnTable(table)
			} else {
				filter = filter.OnJoin(table)
			}
		}
		return &Filter{FieldValueFilter: *filter}
	}

	maps.ProcessOrdered(query, func(key string, value string) {
		if key == "@filter" {
			out.AllOf = append(out.AllOf, FilterFromValue(value))
		} else {
			parts := strings.Split(key, ".")
			switch len(parts) {
			case 1:
				out.AllOf = append(out.AllOf, makeFilter("", parts[0], value))

			default:
				// consider only the last 2 parts of the key.
				parts = parts[len(parts)-2:]
				out.AllOf = append(out.AllOf, makeFilter(parts[0], parts[1], value))
			}
		}
	})

	return out
}
