package context

import (
	"maps"
	"strings"

	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/filtering"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/matching"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// RequestContext represents the filter, field matcher and parameters from an API request.
//
// It will also hold any remaining data from the body (if any).
type RequestContext struct {
	Filter        filtering.Filterer    // vertical filter.
	Matcher       matching.FieldMatcher // horizontal filter.
	Params        map[string]any        // optional parameters for transformations.
	RemainingBody map[string]any        // remaining key/value pairs from the body.
}

// New instantiates a new API request context.
//
// The given query should contain the query parameters from the API request URI.
// The given body should contain the parsed body of the API request (if any).
//
// Special values in the body are either appended to the corresponding query value (field filter)
// or overwrite the corresponding query value (filter & params).
func New(query map[string]string, body map[string]any, primary string) (*RequestContext, error) {
	var fields string
	var params map[string]any
	var filter filtering.Filterer

	if query != nil {
		if fields2, ok := query["@fields"]; ok {
			delete(query, "@fields")
			fields = fields2
		}

		if params2, ok := query["@params"]; ok {
			delete(query, "@params")
			params = convertParams(params2)
		}
	}

	if body != nil {
		if fields2, ok := body["@fields"]; ok {
			delete(body, "@fields")
			fields = strings.Join([]string{fields, convert.AnyToString(fields2)}, ",")
		}

		if params2, ok := body["@params"]; ok {
			delete(body, "@params")

			m := convertParams(params2)
			if params == nil {
				params = m
			} else {
				maps.Insert(params, maps.All(m))
			}
		}

		if filter2, ok := body["@filter"]; ok {
			delete(body, "@filter")

			switch t := filter2.(type) {
			case string:
				filter = filtering.FilterFromAny(primary, t)
			case filtering.Filterer:
				filter = t
			case map[string]any:
				filter = filtering.FilterFromMap(primary, t)
			default:
				filter = filtering.FilterFromAny(primary, convert.AnyToString(filter2))
			}
		}
	}

	if filter == nil {
		// use an optional filter from the query only if the body didn't produce a filter.
		filter = filtering.FilterFromQuery(primary, query)
	}

	return &RequestContext{
		Filter:        filter,
		Matcher:       matching.NewFieldMatcher(fields),
		Params:        params,
		RemainingBody: body,
	}, nil
}

func convertParams(in any) map[string]any {
	if m, ok := in.(map[string]any); ok {
		return m
	}

	s := convert.AnyToString(in)
	out := make(map[string]any)
	if err := json.Unmarshal([]byte(s), &out); err == nil {
		return out
	}

	parts := strings.Split(s, ",")
	for i := range parts {
		kv := strings.Split(parts[i], "=")
		if len(kv) == 2 {
			out[kv[0]] = kv[1]
		}
	}

	return out
}
