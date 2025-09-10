package fiber

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// NewTagsHandler instantiates a tag handler.
func NewTagsHandler(logger *slog.Logger, pap *pap.PAP, authorizer authorization.Authorizer) *TagsHandler {
	return &TagsHandler{logger: logger, pap: pap, authorizer: authorizer}
}

// GetTags retrieves all tags from the PAP.
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

// GetTag retrieves a specific tag from the PAP.
func (h *TagsHandler) GetTag(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, PoliciesVersion)

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var tag string
	if tag, ok, err = h.checkKey(req); !ok {
		return err
	}

	t, _, err2 := h.pap.ReadTag(tag)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if t == nil {
		return h.error(req, fiber.StatusNotFound, tagNotFound)
	}
	return req.JSON(t)
}

// PostTag inserts a new tag into the PAP.
func (h *TagsHandler) PostTag(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, PoliciesVersion)

	user, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var tag string
	if tag, ok, err = h.checkKey(req); !ok {
		return err
	}

	var t *oas.Tag
	if t, ok, err = h.checkBody(req, tag); !ok {
		return err
	}

	prev, lastIndex, err2 := h.pap.ReadTag(tag)
	switch {
	case err2 != nil:
		// no-op
	case prev != nil && !req.QueryBool("forceUpsert"):
		return h.error(req, fiber.StatusConflict, tagExists)
	case prev != nil:
		t, err2 = h.pap.UpdateTag(prev, lastIndex, t, user)
	default:
		t, err2 = h.pap.CreateTag(t, user)
	}

	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}
	return req.Status(fiber.StatusCreated).JSON(t)
}

// PutTag replaces a tag in the PAP.
func (h *TagsHandler) PutTag(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, PoliciesVersion)

	user, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var tag string
	if tag, ok, err = h.checkKey(req); !ok {
		return err
	}

	var t *oas.Tag
	if t, ok, err = h.checkBody(req, tag); !ok {
		return err
	}

	prev, lastIndex, err2 := h.pap.ReadTag(tag)
	switch {
	case err2 != nil:
		// no-op
	case prev == nil && !req.QueryBool("forceUpsert"):
		return h.error(req, fiber.StatusNotFound, tagNotFound)
	case prev == nil:
		t, err2 = h.pap.CreateTag(t, user)
	default:
		t, err2 = h.pap.UpdateTag(prev, lastIndex, t, user)
	}

	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}
	return req.JSON(t)
}

// DeleteTag removes a tag from the PAP.
func (h *TagsHandler) DeleteTag(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, PoliciesVersion)

	user, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var tag string
	if tag, ok, err = h.checkKey(req); !ok {
		return err
	}

	prev, lastIndex, err2 := h.pap.ReadTag(tag)
	switch {
	case err2 != nil:
		// no-op
	case prev == nil && !req.QueryBool("ignoreMissing"):
		return h.error(req, fiber.StatusNotFound, tagNotFound)
	case prev == nil:
		prev = &oas.Tag{Id: tag}
	default:
		prev, err2 = h.pap.DeleteTag(prev, lastIndex, user)
	}

	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}
	return req.JSON(prev)
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

func (h *TagsHandler) checkKey(req *fiber.Ctx) (string, bool, error) {
	id := req.Params("tag")
	if id == "" || len(id) > 40 {
		return "", false, h.error(req, fiber.StatusBadRequest, tagKeyError)
	}

	return id, true, nil
}

func (h *TagsHandler) checkBody(req *fiber.Ctx, tag string) (*oas.Tag, bool, error) {
	var t oas.Tag
	if err := req.BodyParser(&t); err != nil {
		return nil, false, h.error(req, fiber.StatusBadRequest, err)
	}

	var errs []error
	if t.Id != tag {
		if t.Id != "" {
			errs = append(errs, h.error(req, fiber.StatusBadRequest, tagKeyMismatch))
		} else {
			t.Id = tag
		}
	}

	switch {
	case t.Name == "":
		errs = append(errs, errors.New("title must be filled"))
	case len(t.Name) > 80:
		errs = append(errs, errors.New("title too long (max 80 characters)"))
	}

	// TODO: other checks!

	if len(errs) > 0 {
		return nil, false, h.error(req, fiber.StatusBadRequest, errors.Join(errs...))
	}

	return &t, true, nil
}

// TagsHandler implements the interface for handling requests about policy tags.
type TagsHandler struct {
	logger     *slog.Logger
	pap        *pap.PAP
	authorizer authorization.Authorizer
}

var (
	tagNotFound    = errors.New("tag not found")
	tagExists      = errors.New("tag already exists")
	tagKeyError    = errors.New("tag must be filled and less or equal 40 characters")
	tagKeyMismatch = errors.New("tag mismatch")
)
