package filtering

import (
	"regexp"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/maps"
)

// FilterFromQuery returns a filter based on a previously decoded HTTP query string.
func FilterFromQuery(primary string, query map[string]string) *Filter {
	c := newConstructor(primary)

	maps.ProcessOrdered(query, func(key string, value string) {
		if key == "@filter" {
			c.addFilter(FilterFromAny(primary, value))
		} else {
			parts := strings.Split(key, ".")

			switch len(parts) {
			case 1:
				c.addTableKeyValue("", parts[0], value)

			default:
				// consider only the last 2 parts of the key.
				parts = parts[len(parts)-2:]
				c.addTableKeyValue(parts[0], parts[1], value)
			}
		}
	})

	return c.normalizedResult()
}

// FilterFromMap returns a filter based on the given map.
func FilterFromMap(primary string, m map[string]any) *Filter {
	c := newConstructor(primary)

	maps.ProcessOrdered(m, func(key string, value any) {
		if key == "@filter" {
			c.addFilter(FilterFromAny(primary, convert.AnyToString(value)))
		} else {
			parts := strings.Split(key, ".")

			switch len(parts) {
			case 1:
				c.addTableKeyValue("", parts[0], value)

			default:
				// consider only the last 2 parts of the key.
				parts = parts[len(parts)-2:]
				c.addTableKeyValue(parts[0], parts[1], value)
			}
		}
	})

	return c.normalizedResult()
}

// FilterFromString returns a filter based on the given string.
func FilterFromString(primary, in string) *Filter {
	c := newConstructor(primary)

	list := strings.Split(in, ",")
	for i := range list {
		if parts := filterRX.FindStringSubmatch(list[i]); len(parts) == 5 {
			var table, field string
			parts2 := strings.Split(parts[1], ".")
			switch len(parts2) {
			case 1:
				field = parts2[0]
			default:
				table = parts2[len(parts2)-2]
				field = parts2[len(parts2)-1]
			}

			var filter *FieldValueFilter
			if s := parts[4]; s != "" {
				// nil, not nil; so no value, just existence test.
				exists := strings.EqualFold("!nil", s) || strings.EqualFold("not nil", s)
				filter = Exists(field, exists)
			} else {
				// other compare with value.
				filter = Compare(field, enums.CompareTypeFromString(parts[2]), parts[3])
			}

			if table != "" {
				if strings.EqualFold(c.primary, table) {
					filter.OnTable(table)
				} else {
					filter.OnJoin(table)
				}
			}

			c.addFilter(&Filter{FieldValueFilter: *filter})
		}
	}

	return c.normalizedResult()
}

func newConstructor(primary string) *constructor {
	return &constructor{
		primary: primary,
		result:  &Filter{AllOfFilter: AllOfFilter{AllOf: make(Filters, 0)}},
	}
}

func (c *constructor) addFilter(filter *Filter) {
	c.result.AllOf = append(c.result.AllOf, filter)
}

func (c *constructor) addTableKeyValue(table, key string, value any) {
	filter := Compare(key, enums.IsEqual, value)

	if table != "" {
		if strings.EqualFold(c.primary, table) {
			filter = filter.OnTable(table)
		} else {
			filter = filter.OnJoin(table)
		}
	}

	c.result.AllOf = append(c.result.AllOf, &Filter{FieldValueFilter: *filter})
}

func (c *constructor) normalizedResult() *Filter {
	if len(c.result.AllOf) == 1 {
		return c.result.AllOf[0]
	}
	return c.result
}

type constructor struct {
	primary string
	result  *Filter
}

var filterRX = regexp.MustCompile(`(?i:^\s*([a-zA-Z0-9_.-]+)\s*(?:(==|!=|=>|>=|=<|<=|=|<|>|in|not in|!in|like|not like|regex|rx|not regex|not rx)\s*(.+)|(nil|is nil|!nil|not nil)\s*)$)`)
