package fiber

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
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
	logger     *slog.Logger
	cache      *pip.PIP
	authorizer authorization.Authorizer
	principalResolver
}

// NewAttributesHandler instantiates a policy handler.
func NewAttributesHandler(logger *slog.Logger, pip *pip.PIP, authorizer authorization.Authorizer, opts ...HandlerOption) AttributesHandler {
	return &attributesHandler{logger: logger, cache: pip, authorizer: authorizer, principalResolver: newPrincipalResolver(logger, opts)}
}

// GetAttributes implements the AttributesHandler interface.
func (h *attributesHandler) GetAttributes(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	if _, err := h.authorize(req); err != nil {
		return err
	}

	resp := make([]*oas.Attribute, 0, 32)
	h.cache.IterateAttributes(func(attr *models.Attribute) {
		resp = append(resp, attr.ToOAS())
	})

	return h.respond(req, resp)
}

// GetAttribute implements the AttributesHandler interface.
func (h *attributesHandler) GetAttribute(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	if _, err := h.authorize(req); err != nil {
		return err
	}

	key, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	a, _, err2 := h.cache.GetAttribute(key)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if a == nil {
		return h.error(req, fiber.StatusNotFound, errAttrNotFound)
	}

	out := a.ToOAS()

	if audit, err3 := h.cache.GetAttributeAudit(key); err3 != nil {
		h.logger.Warn("failed to read policy audit", "key", key, "err", err3)
	} else {
		out.AuditLog = audit
	}

	if usage, err3 := h.cache.GetAttributeDeployments(key); err3 != nil {
		h.logger.Warn("failed to read policy deployments", "key", key, "err", err3)
	} else {
		out.UsageData = usage
	}

	return h.respond(req, out)
}

// GetAttributeVersions implements the AttributesHandler interface.
func (h *attributesHandler) GetAttributeVersions(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	if _, err := h.authorize(req); err != nil {
		return err
	}

	id, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	list, err2 := h.cache.GetAttributeVersions(id)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	return h.respond(req, list)
}

// GetAttributeVersion implements the AttributesHandler interface.
func (h *attributesHandler) GetAttributeVersion(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	if _, err := h.authorize(req); err != nil {
		return err
	}

	id, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	version, err := h.checkVersion(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	attr, err2 := h.cache.GetAttributeVersion(id, version)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if attr == nil {
		return h.error(req, fiber.StatusNotFound, errAttrNotFound)
	}

	return h.respond(req, attr)
}

// PostAttribute implements the AttributesHandler interface.
func (h *attributesHandler) PostAttribute(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	key, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	a, code, err := h.checkBody(req, key)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	a2, _, err2 := h.cache.GetAttribute(a.Key)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if a2 != nil && !req.QueryBool(ParamForceUpsert) {
		return h.error(req, fiber.StatusConflict, errAttrExists)
	}

	if a2, err2 = h.cache.AddAttributeFromOAS(a, user); err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}
	return h.respond(req.Status(fiber.StatusCreated), a2.ToOAS())
}

// PutAttribute implements the AttributesHandler interface.
func (h *attributesHandler) PutAttribute(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	key, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	a, code, err := h.checkBody(req, key)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	a2, _, err2 := h.cache.GetAttribute(a.Key)
	if err2 != nil {
		h.logger.Error("failed to get attribute", "error", err2)
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if a2 == nil && !req.QueryBool(ParamForceUpsert) {
		return h.error(req, fiber.StatusNotFound, errAttrNotFound)
	}

	if a2, err2 = h.cache.AddAttributeFromOAS(a, user); err2 != nil {
		h.logger.Error("failed to save attribute", "error", err2)
		return h.error(req, fiber.StatusInternalServerError, err2)
	}
	return h.respond(req, a2.ToOAS())
}

// PatchAttributeStatus implements the AttributesHandler interface.
func (h *attributesHandler) PatchAttributeStatus(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	key, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	a, code, err := h.checkBodyStatus(req, key)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	a2, _, err2 := h.cache.GetAttribute(a.Key)
	if err2 != nil {
		h.logger.Error("failed to get attribute", "error", err2)
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if a2, err2 = h.cache.UpdateAttributeStatus(a.Key, models.StatusFromString(a.Status), user); err2 != nil {
		h.logger.Error("failed to save attribute status", "error", err2)
		return h.error(req, fiber.StatusBadRequest, err2)
	}
	return h.respond(req, a2.ToOAS())
}

// PostAttributeRestore implements the AttributesHandler interface.
func (h *attributesHandler) PostAttributeRestore(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	key, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	version, err := h.checkVersion(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	attr, err2 := h.cache.RestoreAttributeVersion(key, version, user)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	return h.respond(req, attr)
}

// DeleteAttribute implements the AttributesHandler interface.
func (h *attributesHandler) DeleteAttribute(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	key, err := h.checkKey(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	a2, _, err2 := h.cache.GetAttribute(key)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if a2 == nil && !req.QueryBool(ParamIgnoreMissing) {
		return h.error(req, fiber.StatusNotFound, errAttrNotFound)
	}

	if a2, err2 = h.cache.RemoveAttribute(key, user); err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}
	return h.respond(req, a2.ToOAS())
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
		checkStatus(a.Status)

	if chk.checkFailed() {
		return nil, chk.firstCode(), chk.error()
	}

	return &a, "", nil
}

func (h *attributesHandler) authorize(req *fiber.Ctx) (identity.Principal, error) {
	return authorizeRequest(h.authorizer, req, h.logger)
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
