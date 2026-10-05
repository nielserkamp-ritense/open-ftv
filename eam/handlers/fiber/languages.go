package fiber

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/management"
)

// NewLanguagesHandler instantiates a policy language handler. Every operation is decided by
// the language service through the authorizer; a nil authorizer permits everything.
func NewLanguagesHandler(logger *slog.Logger, store management.LanguageStore, authorizer authorization.RequestAuthorizer, opts ...HandlerOption) *LanguagesHandler {
	return &LanguagesHandler{
		logger:            logger,
		svc:               management.NewLanguageService(store, authorizer),
		secured:           authorizer != nil,
		principalResolver: newPrincipalResolver(logger, opts),
	}
}

// GetLanguages retrieves the list of supported policy languages.
func (h *LanguagesHandler) GetLanguages(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	resp, err := h.svc.List(req.UserContext(), callerOf(req, h.secured, h.logger))
	if err != nil {
		if errors.Is(err, authorization.ErrForbidden) {
			return auth.Forbidden(req, err, h.logger)
		}

		return internalError(req, h.logger, err)
	}

	return h.respond(req, resp)
}

// LanguagesHandler implements the interface for handling requests about policy languages.
type LanguagesHandler struct {
	logger  *slog.Logger
	svc     *management.LanguageService
	secured bool // an authorizer decides; the routes must carry the Identify middleware.
	principalResolver
}
