package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
)

// NewLanguagesHandler instantiates a policy language handler.
func NewLanguagesHandler(logger *slog.Logger, pap *pap.PAP, authorizer authorization.Authorizer, opts ...HandlerOption) *LanguagesHandler {
	return &LanguagesHandler{logger: logger, pap: pap, authorizer: authorizer, principalResolver: newPrincipalResolver(logger, opts)}
}

// GetLanguages retrieves the list of supported policy languages.
func (h *LanguagesHandler) GetLanguages(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	if _, err := h.authorize(req); err != nil {
		return err
	}

	resp, err2 := h.pap.ListLanguages()
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	return h.respond(req, resp)
}

func (h *LanguagesHandler) authorize(req *fiber.Ctx) (identity.Principal, error) {
	return authorizeRequest(h.authorizer, req, h.logger)
}

func (h *LanguagesHandler) error(req *fiber.Ctx, status int, err error) error {
	h.logger.Error("request error", "path", req.Path(), "err", err, "status", status)
	return server.SendMessageResponse(req, status, err.Error())
}

// LanguagesHandler implements the interface for handling requests about policy languages.
type LanguagesHandler struct {
	logger     *slog.Logger
	pap        *pap.PAP
	authorizer authorization.Authorizer
	principalResolver
}
