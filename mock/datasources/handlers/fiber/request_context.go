package fiber

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/context"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/decode"
)

func buildRequestContext(req *fiber.Ctx, primary string) (*context.RequestContext, error) {
	query := req.Queries()

	body, err := decodeBody(req)
	if err != nil {
		return nil, err
	}

	return context.New(query, body, primary)
}

func decodeBody(req *fiber.Ctx) (map[string]any, error) {
	ct := req.Get(fiber.HeaderContentEncoding)
	if ct == "" {
		ct = fiber.MIMEApplicationJSON
	}

	var body map[string]any
	if data := req.Body(); len(data) > 0 {
		var err error
		if body, err = decode.ParseBody(data, ct); err != nil {
			return nil, err
		}
	}

	return body, nil
}

func pkValues(req *fiber.Ctx, keys []string) []any {
	out := make([]any, len(keys))
	for i := range keys {
		out[i] = req.Params(keys[i])
	}
	return out
}

// pathHasKeys reports whether path declares a named parameter for every one of keys,
// e.g. "/aanvraag/:id" satisfies key "id" but the collection path "/aanvragen" does not.
func pathHasKeys(path string, keys []string) bool {
	params := make(map[string]bool)

	for _, segment := range strings.Split(path, "/") {
		if name, ok := strings.CutPrefix(segment, ":"); ok {
			params[strings.TrimSuffix(name, "?")] = true
		}
	}

	for _, key := range keys {
		if !params[key] {
			return false
		}
	}

	return true
}
