package context

import (
	"maps"
	"strings"

	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/filtering"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/matching"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
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
func New(query map[string]string, body map[string]any, ds *schema.Datasource, primary string) (*RequestContext, error) {
	var fields string
	var params map[string]any

	if query != nil {
		fields = query["@fields"]
		delete(query, "@fields")

		params = convertParams(query["@params"])
		delete(query, "@params")
	}

	if body != nil {
		if fields2, ok := body["@fields"]; ok {
			fields = strings.Join([]string{fields, convert.AnyToString(fields2)}, ",")
			delete(body, "@fields")
		}

		if params2, ok := body["@params"]; ok {
			m := convertParams(convert.AnyToString(params2))
			maps.Insert(params, maps.All(m))
			delete(body, "@params")
		}

		if filter, ok := body["@filter"]; ok {
			if m2, ok2 := filter.(map[string]any); ok2 {
				b, _ := json.Marshal(m2)
				query["@filter"] = string(b)
			} else {
				query["@filter"] = convert.AnyToString(filter)
			}
			delete(body, "@filter")
		}

	}

	return &RequestContext{
		Filter:        filtering.FilterFromQuery(ds, primary, query),
		Matcher:       matching.NewFieldMatcher(fields),
		Params:        params,
		RemainingBody: body,
	}, nil
}

func convertParams(in string) map[string]any {
	out := make(map[string]any)
	if err := json.Unmarshal([]byte(in), &out); err == nil {
		return out
	}

	parts := strings.Split(in, ",")
	for i := range parts {
		kv := strings.Split(parts[i], "=")
		if len(kv) == 2 {
			out[kv[0]] = kv[1]
		}
	}

	return out
}
