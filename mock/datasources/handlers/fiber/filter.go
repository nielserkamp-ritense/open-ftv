package fiber

import (
	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/types"
)

func buildFilter(req *fiber.Ctx) map[string]any {
	q := req.Queries()
	out := make(map[string]any, len(q))
	for k, v := range q {
		out[k] = v
	}
	return out
}

func extractFieldMatcher(filter map[string]any) (map[string]any, types.FieldMatcher) {
	expr, _ := filter["fields"].(string)
	delete(filter, "fields")
	matcher := types.NewFieldMatcher(expr)
	return filter, matcher
}
