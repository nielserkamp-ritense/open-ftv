package fiber

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

var (
	errTagNotFound    = errors.New("tag not found")
	errTagExists      = errors.New("tag already exists")
	errTagConcurrency = errors.New("tag was modified by another request")
	errTagKeyError    = errors.New("tag must be filled and less or equal 40 characters")

	errTagMismatch = checkIssue{code: "E05005", msg: "tag mismatch"}
)

// TagsHandler implements the interface for handling requests about policy tags.
type TagsHandler struct {
	logger     *slog.Logger
	pap        *pap.PAP
	authorizer authorization.Authorizer
	principalResolver
}

// NewTagsHandler instantiates a tag handler.
func NewTagsHandler(logger *slog.Logger, pap *pap.PAP, authorizer authorization.Authorizer, opts ...HandlerOption) *TagsHandler {
	return &TagsHandler{logger: logger, pap: pap, authorizer: authorizer, principalResolver: newPrincipalResolver(logger, opts)}
}

// GetTags retrieves all tags from the PAP.
func (h *TagsHandler) GetTags(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	if _, err := h.authorize(req); err != nil {
		return err
	}

	resp, err2 := h.pap.ListTags()
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	return h.respond(req, resp)
}

// GetTag retrieves a specific tag from the PAP.
func (h *TagsHandler) GetTag(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	if _, err := h.authorize(req); err != nil {
		return err
	}

	tag, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	t, _, err2 := h.pap.ReadTag(tag)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if t == nil {
		return h.error(req, fiber.StatusNotFound, errTagNotFound)
	}
	return h.respond(req, t)
}

// PostTag inserts a new tag into the PAP.
func (h *TagsHandler) PostTag(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	tag, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	t, code, err := h.checkBody(req, tag)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	prev, lastIndex, err2 := h.pap.ReadTag(tag)
	switch {
	case err2 != nil:
		// no-op
	case prev != nil && !req.QueryBool("forceUpsert"):
		return h.error(req, fiber.StatusConflict, errTagExists)
	case prev != nil:
		t, err2 = h.pap.UpdateTag(prev, lastIndex, t, user)
	default:
		t, err2 = h.pap.CreateTag(t, user)
	}

	if err2 != nil {
		return h.tagConcurrencyError(req, err2)
	}
	return h.respond(req.Status(fiber.StatusCreated), t)
}

// PutTag replaces a tag in the PAP.
func (h *TagsHandler) PutTag(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	tag, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	t, code, err := h.checkBody(req, tag)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	prev, lastIndex, err2 := h.pap.ReadTag(tag)
	switch {
	case err2 != nil:
		// no-op
	case prev == nil && !req.QueryBool("forceUpsert"):
		return h.error(req, fiber.StatusNotFound, errTagNotFound)
	case prev == nil:
		t, err2 = h.pap.CreateTag(t, user)
	default:
		t, err2 = h.pap.UpdateTag(prev, lastIndex, t, user)
	}

	if err2 != nil {
		return h.tagConcurrencyError(req, err2)
	}
	return h.respond(req, t)
}

// DeleteTag removes a tag from the PAP.
func (h *TagsHandler) DeleteTag(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	tag, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	prev, lastIndex, err2 := h.pap.ReadTag(tag)
	switch {
	case err2 != nil:
		// no-op
	case prev == nil && !req.QueryBool("ignoreMissing"):
		return h.error(req, fiber.StatusNotFound, errTagNotFound)
	case prev == nil:
		prev = &oas.Tag{Id: tag}
	default:
		prev, err2 = h.pap.DeleteTag(prev, lastIndex, user)
	}

	if err2 != nil {
		return h.tagConcurrencyError(req, err2)
	}
	return h.respond(req, prev)
}

func (h *TagsHandler) authorize(req *fiber.Ctx) (identity.Principal, error) {
	return authorizeRequest(h.authorizer, req, h.logger)
}

func (h *TagsHandler) tagConcurrencyError(req *fiber.Ctx, err error) error {
	if errors.Is(err, pap.ErrTagConcurrency) {
		return h.error(req, fiber.StatusConflict, errTagConcurrency)
	}
	return h.error(req, fiber.StatusInternalServerError, err)
}

func (h *TagsHandler) error(req *fiber.Ctx, status int, err error) error {
	h.logger.Error("request error", "path", req.Path(), "err", err, "status", status)
	return server.SendMessageResponse(req, status, err.Error())
}

// badRequest logs a warning and returns a 400 with the given validation error's code and message.
func (h *TagsHandler) badRequest(req *fiber.Ctx, code string, err error) error {
	h.logger.Warn("tag request rejected", "path", req.Path(), "err", err, "status", fiber.StatusBadRequest)
	return server.SendProblemResponse(req, fiber.StatusBadRequest, code, err.Error())
}

func (h *TagsHandler) checkKey(req *fiber.Ctx) (string, error) {
	id := req.Params("tag")
	if id == "" || len(id) > 40 {
		return "", errTagKeyError
	}

	return id, nil
}

// checkBody parses and validates the request body.
func (h *TagsHandler) checkBody(req *fiber.Ctx, tag string) (*oas.Tag, string, error) {
	var t oas.Tag
	if err := req.BodyParser(&t); err != nil {
		return nil, codeBadRequest, err
	}

	chk := newFieldChecker().
		checkIdentifiers(tag, &t.Id, errTagMismatch).
		checkTitle(t.Name)

	if chk.checkFailed() {
		return nil, chk.firstCode(), chk.error()
	}

	return &t, "", nil
}
