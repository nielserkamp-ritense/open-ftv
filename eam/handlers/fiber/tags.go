package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
)

// NewTagsHandler instantiates a tag handler.
func NewTagsHandler(logger *slog.Logger, pap *pap.PAP, authorizer authorization.Authorizer) *TagsHandler {
	return &TagsHandler{logger: logger, pap: pap, authorizer: authorizer}
}

// GetTags retrieves the list of known tags.
func (h *TagsHandler) GetTags(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, PoliciesVersion)

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	resp, err2 := h.pap.ListTags()
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	return req.JSON(resp)
}

func (h *TagsHandler) authorize(req *fiber.Ctx) (string, bool, error) {
	if h.authorizer == nil {
		return auth.SystemUser, true, nil
	}

	resp, err := h.authorizer.Authorize(auth.FormatRequest(req))

	// TODO: log authorization decision to auth-decision log.

	return auth.Check(req, resp, err, h.logger)
}

func (h *TagsHandler) error(req *fiber.Ctx, status int, err error) error {
	h.logger.Error("request error", "path", req.Path(), "err", err, "status", status)
	return server.SendMessageResponse(req, status, err.Error())
}

// TagsHandler implements the interface for handling requests about policy tags.
type TagsHandler struct {
	logger     *slog.Logger
	pap        *pap.PAP
	authorizer authorization.Authorizer
}
