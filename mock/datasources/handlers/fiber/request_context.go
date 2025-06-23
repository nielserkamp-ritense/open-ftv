package fiber

import (
	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/context"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/decode"
)

func buildRequestContext(req *fiber.Ctx, primary string) (*context.RequestContext, error) {
	query := req.Queries()

	var body map[string]any
	if data := req.Body(); len(data) > 0 {
		var err error
		if body, err = decode.ParseBody(data, req.Get(fiber.HeaderContentEncoding)); err != nil {
			return nil, err
		}
	}

	return context.New(query, body, primary)
}
