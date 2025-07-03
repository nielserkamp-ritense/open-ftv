package filtering

import (
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
)

// NewTableFilter reduces the filter to the fields that exist in the given table.
func NewTableFilter(table *schema.Table, filter map[string]any) map[string]any {
	if len(filter) == 0 {
		return filter
	}

	test := func(id string) bool {
		if !strings.EqualFold(id, "fields") {
			for _, field := range table.Fields {
				if strings.EqualFold(id, field.ID) {
					return true
				}
			}
		}
		return false
	}

	out := make(map[string]any, len(filter))

	for k, v := range filter {
		parts := strings.Split(k, ".")
		switch len(parts) {
		case 1:
			// unqualified.
			if test(parts[0]) {
				out[parts[0]] = v
			}
		case 2:
			// qualified by table id.
			if strings.EqualFold(parts[0], table.ID) {
				if test(parts[1]) {
					out[parts[1]] = v
				}
			}
		}
	}

	return out
}
