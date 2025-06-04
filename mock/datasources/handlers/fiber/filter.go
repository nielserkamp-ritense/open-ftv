package fiber

import (
	"strings"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/filtering"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/matching"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/decode"
)

func buildFilters(req *fiber.Ctx, ds *schema.Datasource) (filtering.Filterer, matching.FieldMatcher) {
	query := req.Queries()
	fields := query["@fields"]
	delete(query, "@fields")

	if body := req.Body(); len(body) > 0 {
		if m, err := decode.ParseBody(body, req.Get(fiber.HeaderContentEncoding)); err == nil {
			if fields2, ok := m["@fields"]; ok {
				fields = strings.Join([]string{fields, convert.AnyToString(fields2)}, ",")
			}

			if filter, ok := m["@filter"]; ok {
				if m2, ok2 := filter.(map[string]any); ok2 {
					b, _ := json.Marshal(m2)
					query["@filter"] = string(b)
				} else {
					query["@filter"] = convert.AnyToString(filter)
				}
			}
		}
	}

	return filtering.FilterFromQuery(ds, query), matching.NewFieldMatcher(fields)
}
