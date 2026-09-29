package fiber

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/management"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

// AttributesVersion is the full semantic API version for the attribute endpoints.
const AttributesVersion = "1.7.1" // check against oas/attributes/openapi.yaml!

var (
	errAttrNotFound     = errors.New("attribute not found")
	errAttrExists       = errors.New("attribute already exists")
	errAttrKeyError     = errors.New("attribute key must be filled and less or equal 500 characters")
	errAttrVersionError = errors.New("attribute version must be filled and positive")

	errAttributeKeyMismatch = checkIssue{code: "E04005", msg: "attribute key mismatch"}
)

// AttributesHandler represents the interface for handling requests about attributes.
type AttributesHandler interface {
	GetAttributes(req *fiber.Ctx) error
	GetAttribute(req *fiber.Ctx) error
	GetAttributeVersions(req *fiber.Ctx) error
	GetAttributeVersion(req *fiber.Ctx) error
	PostAttribute(req *fiber.Ctx) error
	PutAttribute(req *fiber.Ctx) error
	PatchAttributeStatus(req *fiber.Ctx) error
	PostAttributeRestore(req *fiber.Ctx) error
	DeleteAttribute(req *fiber.Ctx) error
}

type attributesHandler struct {
	logger  *slog.Logger
	svc     *management.AttributeService
	secured bool // an authorizer decides; the routes must carry the Identify middleware.
	principalResolver
}

// NewAttributesHandler instantiates an attribute handler. Every operation is decided by the
// attribute service through the authorizer; a nil authorizer permits everything. *pip.PIP is
// an AttributeStore.
func NewAttributesHandler(logger *slog.Logger, store management.AttributeStore, authorizer authorization.RequestAuthorizer, opts ...HandlerOption) AttributesHandler {
	return &attributesHandler{
		logger:            logger,
		svc:               management.NewAttributeService(logger, store, authorizer),
		secured:           authorizer != nil,
		principalResolver: newPrincipalResolver(logger, opts),
	}
}

// caller is the request-scoped Principal this request runs as.
func (h *attributesHandler) caller(req *fiber.Ctx) *authorization.RequestPrincipal {
	return callerOf(req, h.secured, h.logger)
}

// GetAttributes implements the AttributesHandler interface.
func (h *attributesHandler) GetAttributes(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	list, err := h.svc.List(req.UserContext(), h.caller(req))
	if err != nil {
		return h.fail(req, err)
	}

	resp := make([]*oas.Attribute, len(list))
	for i := range list {
		resp[i] = list[i].ToOAS()
	}

	return h.respond(req, resp)
}

// GetAttribute implements the AttributesHandler interface.
func (h *attributesHandler) GetAttribute(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	key, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	details, err := h.svc.Get(req.UserContext(), h.caller(req), key)
	if err != nil {
		return h.fail(req, err)
	}

	out := details.Attribute.ToOAS()
	out.AuditLog = details.AuditLog
	out.UsageData = details.UsageData

	return h.respond(req, out)
}

// GetAttributeVersions implements the AttributesHandler interface.
func (h *attributesHandler) GetAttributeVersions(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	key, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	list, err := h.svc.Versions(req.UserContext(), h.caller(req), key)
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, list)
}

// GetAttributeVersion implements the AttributesHandler interface.
func (h *attributesHandler) GetAttributeVersion(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	key, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	version, err := h.checkVersion(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	attr, err := h.svc.Version(req.UserContext(), h.caller(req), key, version)
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, attr)
}

// PostAttribute implements the AttributesHandler interface.
func (h *attributesHandler) PostAttribute(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	key, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	a, code, err := h.checkBody(req, key)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	out, err := h.svc.Create(req.UserContext(), h.caller(req), a, req.QueryBool(ParamForceUpsert))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req.Status(fiber.StatusCreated), out.ToOAS())
}

// PutAttribute implements the AttributesHandler interface.
func (h *attributesHandler) PutAttribute(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	key, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	a, code, err := h.checkBody(req, key)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	out, err := h.svc.Update(req.UserContext(), h.caller(req), a, req.QueryBool(ParamForceUpsert))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, out.ToOAS())
}

// PatchAttributeStatus implements the AttributesHandler interface.
func (h *attributesHandler) PatchAttributeStatus(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	key, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	a, code, err := h.checkBodyStatus(req, key)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	out, err := h.svc.SetStatus(req.UserContext(), h.caller(req), key, models.StatusFromString(a.Status))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, out.ToOAS())
}

// PostAttributeRestore implements the AttributesHandler interface.
func (h *attributesHandler) PostAttributeRestore(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	key, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	version, err := h.checkVersion(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	out, err := h.svc.Restore(req.UserContext(), h.caller(req), key, version)
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, out)
}

// DeleteAttribute implements the AttributesHandler interface.
func (h *attributesHandler) DeleteAttribute(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	key, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	out, err := h.svc.Delete(req.UserContext(), h.caller(req), key, req.QueryBool(ParamIgnoreMissing))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, out.ToOAS())
}

func (h *attributesHandler) checkKey(req *fiber.Ctx) (string, error) {
	key := req.Params("key")
	if key == "" || len(key) > 200 {
		return "", errAttrKeyError
	}

	return key, nil
}

func (h *attributesHandler) checkVersion(req *fiber.Ctx) (int, error) {
	v, err := req.ParamsInt("version")
	if err != nil {
		return 0, err
	}

	if v <= 0 {
		return 0, errAttrVersionError
	}

	return v, nil
}

// checkBody parses and validates the request body, returning the Code and error to report via
// badRequest on failure. A nil error means the returned Attribute is valid.
func (h *attributesHandler) checkBody(req *fiber.Ctx, key string) (*oas.Attribute, string, error) {
	var a oas.Attribute
	if err := req.BodyParser(&a); err != nil {
		return nil, codeBadRequest, err
	}

	chk := newFieldChecker().
		checkIdentifiers(key, &a.Key, errAttributeKeyMismatch).
		checkStatus(a.Status).
		checkAttrType(a.Type).
		checkTitle(a.Metadata.Title).
		checkTags(a.Metadata.Tags)

	if chk.checkFailed() {
		return nil, chk.firstCode(), chk.error()
	}

	return &a, "", nil
}

// checkBodyStatus is like checkBody but for the status-only request body.
func (h *attributesHandler) checkBodyStatus(req *fiber.Ctx, key string) (*oas.AttributeStatus, string, error) {
	var a oas.AttributeStatus
	if err := req.BodyParser(&a); err != nil {
		return nil, codeBadRequest, err
	}

	chk := newFieldChecker().
		checkIdentifiers(key, &a.Key, errAttributeKeyMismatch).
		requireStatus(a.Status)

	if chk.checkFailed() {
		return nil, chk.firstCode(), chk.error()
	}

	return &a, "", nil
}

// fail answers a failed service call: a denial is 403 regardless of whether the attribute
// exists, an invalid status transition 400, a missing attribute 404, an existing one 409, and
// anything else (including a permitted caller that could not be recorded) 500 without the cause.
func (h *attributesHandler) fail(req *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, authorization.ErrForbidden):
		return auth.Forbidden(req, err, h.logger)
	case errors.Is(err, management.ErrInvalidTransition):
		return h.error(req, fiber.StatusBadRequest, err)
	case errors.Is(err, management.ErrNotFound):
		return h.error(req, fiber.StatusNotFound, errAttrNotFound)
	case errors.Is(err, management.ErrExists):
		return h.error(req, fiber.StatusConflict, errAttrExists)
	default:
		return internalError(req, h.logger, err)
	}
}

func (h *attributesHandler) error(req *fiber.Ctx, status int, err error) error {
	h.logger.Error("request error", "path", req.Path(), "err", err, "status", status)
	return server.SendMessageResponse(req, status, err.Error())
}

// badRequest logs a warning and returns a 400 with the given validation error's code and message.
func (h *attributesHandler) badRequest(req *fiber.Ctx, code string, err error) error {
	h.logger.Warn("attribute request rejected", "path", req.Path(), "err", err, "status", fiber.StatusBadRequest)
	return server.SendProblemResponse(req, fiber.StatusBadRequest, code, err.Error())
}
