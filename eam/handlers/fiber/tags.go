package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	authRequest "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// NewTagsHandler instantiates a tag handler.
func NewTagsHandler(logger *slog.Logger, tags []*policies.Tag, authorizer authorization.Authorizer) *TagsHandler {
	return &TagsHandler{logger: logger, tags: tags, authorizer: authorizer}
}

// GetTags retrieves the list of known tags.
func (h *TagsHandler) GetTags(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, PoliciesVersion)

	user, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	_ = user

	return req.JSON(h.tags)
}

func (h *TagsHandler) authorize(req *fiber.Ctx) (string, bool, error) {
	if h.authorizer == nil {
		return authRequest.SystemUser, true, nil
	}

	resp, err := h.authorizer.Authorize(authRequest.FormatRequest(req))

	// TODO: log authorization decision to auth-decision log.

	return authRequest.Check(req, resp, err, h.logger)
}

// TagsHandler implements the interface for handling requests about policy tags.
type TagsHandler struct {
	logger     *slog.Logger
	tags       []*policies.Tag
	authorizer authorization.Authorizer
}
