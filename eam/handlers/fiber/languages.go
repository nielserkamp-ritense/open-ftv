package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	authRequest "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// NewLanguagesHandler instantiates a policy language handler.
func NewLanguagesHandler(logger *slog.Logger, authorizer authorization.Authorizer) *LanguagesHandler {
	return &LanguagesHandler{logger: logger, authorizer: authorizer}
}

// GetLanguages retrieves the list of supported policy languages.
func (h *LanguagesHandler) GetLanguages(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, PoliciesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	languages := []models.Language{
		models.REGO,
		models.CEDAR,
		models.CERBOS,
		models.OPENFGA,
	}

	resp := make([]*policies.Language, 0, len(languages))
	for _, l := range languages {
		resp = append(resp, &policies.Language{
			Id:   l.Language(),
			Name: l.String(),
		})
	}
	return req.JSON(resp)
}

func (h *LanguagesHandler) authorize(req *fiber.Ctx) (bool, error) {
	if h.authorizer == nil {
		return true, nil
	}

	resp, err := h.authorizer.Authorize(authRequest.FormatRequest(req))

	// TODO: log authorization decision to auth-decision log.

	return authRequest.Check(req, resp, err)
}

// LanguagesHandler implements the interface for handling requests about policy languages.
type LanguagesHandler struct {
	logger     *slog.Logger
	authorizer authorization.Authorizer
}
