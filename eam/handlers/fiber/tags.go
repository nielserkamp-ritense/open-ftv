package fiber

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/management"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

var (
	errTagNotFound    = errors.New("tag not found")
	errTagExists      = errors.New("tag already exists")
	errTagConcurrency = errors.New("tag was modified by another request")
	errTagKeyError    = errors.New("tag must be filled and less or equal 40 characters")
	errTagInternal    = errors.New("internal error")
	errTagNotRecorded = errors.New("failed to record principal")

	errTagMismatch = checkIssue{code: "E05005", msg: "tag mismatch"}
)

// TagsHandler implements the interface for handling requests about policy tags.
type TagsHandler struct {
	logger  *slog.Logger
	svc     *management.TagService
	secured bool // an authorizer decides; the routes must carry the Identify middleware.
	principalResolver
}

// NewTagsHandler instantiates a tag handler. Every operation is decided by the tag service
// through the authorizer; a nil authorizer permits everything.
func NewTagsHandler(logger *slog.Logger, store *pap.PAP, authorizer authorization.RequestAuthorizer, opts ...HandlerOption) *TagsHandler {
	return &TagsHandler{
		logger:            logger,
		svc:               management.NewTagService(logger, store, authorizer),
		secured:           authorizer != nil,
		principalResolver: newPrincipalResolver(logger, opts),
	}
}

// caller is the request-scoped Principal this request runs as.
func (h *TagsHandler) caller(req *fiber.Ctx) *authorization.RequestPrincipal {
	return callerOf(req, h.secured, h.logger)
}

// GetTags retrieves all tags from the PAP.
func (h *TagsHandler) GetTags(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	list, err := h.svc.List(req.UserContext(), h.caller(req))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, list)
}

// GetTag retrieves a specific tag from the PAP.
func (h *TagsHandler) GetTag(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	tag, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	t, err := h.svc.Get(req.UserContext(), h.caller(req), tag)
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, t)
}

// PostTag inserts a new tag into the PAP.
func (h *TagsHandler) PostTag(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	tag, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	t, code, err := h.checkBody(req, tag)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	t, err = h.svc.Create(req.UserContext(), h.caller(req), t, req.QueryBool("forceUpsert"))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req.Status(fiber.StatusCreated), t)
}

// PutTag replaces a tag in the PAP.
func (h *TagsHandler) PutTag(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	tag, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	t, code, err := h.checkBody(req, tag)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	t, err = h.svc.Update(req.UserContext(), h.caller(req), t, req.QueryBool("forceUpsert"))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, t)
}

// DeleteTag removes a tag from the PAP.
func (h *TagsHandler) DeleteTag(req *fiber.Ctx) error {
	req.Set(HeaderVersion, PoliciesVersion)

	tag, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	t, err := h.svc.Delete(req.UserContext(), h.caller(req), tag, req.QueryBool("ignoreMissing"))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, t)
}

// fail answers a failed service call: a denial is 403 regardless of whether the tag exists, a
// missing tag 404, an existing one or a concurrent change 409, and anything else (including a
// permitted caller that could not be recorded) 500 without the cause.
func (h *TagsHandler) fail(req *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, authorization.ErrForbidden):
		return auth.Forbidden(req, err, h.logger)
	case errors.Is(err, management.ErrNotFound):
		return h.error(req, fiber.StatusNotFound, errTagNotFound)
	case errors.Is(err, management.ErrExists):
		return h.error(req, fiber.StatusConflict, errTagExists)
	case errors.Is(err, pap.ErrTagConcurrency):
		return h.error(req, fiber.StatusConflict, errTagConcurrency)
	default:
		return h.internal(req, err)
	}
}

// internal answers a failure the caller cannot act on with 500. The cause is logged and
// never sent: a store or PDP error names hosts, tables and constraints.
func (h *TagsHandler) internal(req *fiber.Ctx, err error) error {
	h.logger.Error("request error", "path", req.Path(), "err", err, "status", fiber.StatusInternalServerError)

	var notRecorded *authorization.ErrPrincipalNotRecorded
	if errors.As(err, &notRecorded) {
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, errTagNotRecorded.Error())
	}

	return server.SendMessageResponse(req, fiber.StatusInternalServerError, errTagInternal.Error())
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
