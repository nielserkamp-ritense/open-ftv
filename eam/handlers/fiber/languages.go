package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// NewLanguagesHandler instantiates a policy language handler.
func NewLanguagesHandler(logger *slog.Logger, authorizer authorization.Authorizer) *LanguagesHandler {
	return &LanguagesHandler{logger: logger, authorizer: authorizer}
}

// GetLanguages retrieves the list of supported policy languages.
func (h *LanguagesHandler) GetLanguages(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, PoliciesVersion)

	user, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	_ = user

	languages := []models.Language{
		models.REGO,
		models.CEDAR,
		models.CERBOS,
		models.OPENFGA,
	}

	resp := make([]*oas.Language, 0, len(languages))
	for _, l := range languages {
		resp = append(resp, &oas.Language{
			Id:   l.Language(),
			Name: l.String(),
		})
	}
	return req.JSON(resp)
}

func (h *LanguagesHandler) authorize(req *fiber.Ctx) (string, bool, error) {
	if h.authorizer == nil {
		return auth.SystemUser, true, nil
	}

	resp, err := h.authorizer.Authorize(auth.FormatRequest(req))

	// TODO: log authorization decision to auth-decision log.

	return auth.Check(req, resp, err, h.logger)
}

// LanguagesHandler implements the interface for handling requests about policy languages.
type LanguagesHandler struct {
	logger     *slog.Logger
	authorizer authorization.Authorizer
}
