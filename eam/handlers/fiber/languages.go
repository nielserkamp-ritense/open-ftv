package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
)

// NewLanguagesHandler instantiates a policy language handler.
func NewLanguagesHandler(logger *slog.Logger, pap *pap.PAP, authorizer authorization.Authorizer) *LanguagesHandler {
	return &LanguagesHandler{logger: logger, pap: pap, authorizer: authorizer}
}

// GetLanguages retrieves the list of supported policy languages.
func (h *LanguagesHandler) GetLanguages(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	resp, err2 := h.pap.ListLanguages()
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	return req.JSON(resp)
}

func (h *LanguagesHandler) authorize(req *fiber.Ctx) (string, bool, error) {
	if h.authorizer == nil {
		return auth.SystemUser, true, nil
	}

	resp, err := h.authorizer.Authorize(auth.FormatRequest(req))
	return auth.Check(req, resp, err, h.logger)
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
}
