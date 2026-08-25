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

// EntitiesVersion is the full semantic API version for the entity endpoints.
const EntitiesVersion = AttributesVersion

var (
	errEntityNotFound     = errors.New("entity not found")
	errEntityExists       = errors.New("entity already exists")
	errEntityTypeError    = errors.New("entity type must be filled and less or equal 80 characters")
	errEntityIDError      = errors.New("entity ID must be filled and less or equal 200 characters")
	errEntityVersionError = errors.New("entity version must be filled and positive")

	errEntityTypeMismatch = checkIssue{code: "E03005", msg: "entity type mismatch"}
	errEntityIDMismatch   = checkIssue{code: "E03010", msg: "entity id mismatch"}
)

// EntitiesHandler represents the interface for handling requests about entities.
type EntitiesHandler interface {
	GetEntities(req *fiber.Ctx) error
	GetEntity(req *fiber.Ctx) error
	GetEntityVersions(req *fiber.Ctx) error
	GetEntityVersion(req *fiber.Ctx) error
	PostEntity(req *fiber.Ctx) error
	PutEntity(req *fiber.Ctx) error
	PatchEntityStatus(req *fiber.Ctx) error
	PostEntityRestore(req *fiber.Ctx) error
	DeleteEntity(req *fiber.Ctx) error
}

type entitiesHandler struct {
	logger     *slog.Logger
	cache      *pip.PIP
	authorizer authorization.Authorizer
	principalResolver
}

// NewEntitiesHandler instantiates a policy handler.
func NewEntitiesHandler(logger *slog.Logger, store *pip.PIP, authorizer authorization.Authorizer, opts ...HandlerOption) EntitiesHandler {
	return &entitiesHandler{logger: logger, cache: store, authorizer: authorizer, principalResolver: newPrincipalResolver(logger, opts)}
}

// GetEntities implements the EntitiesHandler interface.
func (h *entitiesHandler) GetEntities(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	if _, err := h.authorize(req); err != nil {
		return err
	}

	resp := make([]*oas.Entity, 0, 32)
	h.cache.IterateEntities(func(e *models.Entity) {
		resp = append(resp, e.ToOAS())
	})

	return h.respond(req, &resp)
}

// GetEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) GetEntity(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	if _, err := h.authorize(req); err != nil {
		return err
	}

	ns, id, err := h.checkUID(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	uid := models.EntityUID(ns, id)

	e, _, err2 := h.cache.GetEntity(uid)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if e == nil {
		return h.error(req, fiber.StatusNotFound, errEntityNotFound)
	}

	out := e.ToOAS()

	if audit, err3 := h.cache.GetAttributeAudit(uid); err3 != nil {
		h.logger.Warn("failed to read policy audit", "type", ns, "id", id, "err", err3)
	} else {
		out.AuditLog = audit
	}

	if usage, err3 := h.cache.GetAttributeDeployments(uid); err3 != nil {
		h.logger.Warn("failed to read policy deployments", "type", ns, "id", id, "err", err3)
	} else {
		out.UsageData = usage
	}

	return h.respond(req, out)
}

// GetEntityVersions implements the EntitiesHandler interface.
func (h *entitiesHandler) GetEntityVersions(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	if _, err := h.authorize(req); err != nil {
		return err
	}

	ns, id, err := h.checkUID(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	list, err2 := h.cache.GetEntityVersions(models.EntityUID(ns, id))
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	return h.respond(req, list)
}

// GetEntityVersion implements the EntitiesHandler interface.
func (h *entitiesHandler) GetEntityVersion(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	if _, err := h.authorize(req); err != nil {
		return err
	}

	ns, id, err := h.checkUID(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	version, err := h.checkVersion(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	e, err2 := h.cache.GetEntityVersion(models.EntityUID(ns, id), version)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if e == nil {
		return h.error(req, fiber.StatusNotFound, errEntityNotFound)
	}

	return h.respond(req, e)
}

// PostEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) PostEntity(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	ns, id, err := h.checkUID(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	e1, code, err := h.checkBody(req, ns, id)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	prev, _, err2 := h.cache.GetEntity(models.EntityUID(e1.Type, e1.Id))
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if prev != nil && !req.QueryBool(ParamForceUpsert) {
		return h.error(req, fiber.StatusConflict, errEntityExists)
	}

	e2, err3 := h.cache.AddEntityFromOAS(e1, user)
	if err3 != nil {
		return h.error(req, fiber.StatusInternalServerError, err3)
	}

	return h.respond(req.Status(fiber.StatusCreated), e2.ToOAS())
}

// PutEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) PutEntity(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	ns, id, err := h.checkUID(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	e1, code, err := h.checkBody(req, ns, id)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	prev, _, err2 := h.cache.GetEntity(models.EntityUID(e1.Type, e1.Id))
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if prev == nil && !req.QueryBool(ParamForceUpsert) {
		return h.error(req, fiber.StatusNotFound, errEntityNotFound)
	}

	e2, err3 := h.cache.AddEntityFromOAS(e1, user)
	if err3 != nil {
		return h.error(req, fiber.StatusInternalServerError, err3)
	}

	return h.respond(req, e2.ToOAS())
}

// PatchEntityStatus implements the EntitiesHandler interface.
func (h *entitiesHandler) PatchEntityStatus(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	ns, id, err := h.checkUID(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	e1, code, err := h.checkBodyStatus(req, ns, id)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	uid := models.EntityUID(e1.Type, e1.Id)

	e2, _, err2 := h.cache.GetEntity(uid)
	if err2 != nil {
		h.logger.Error("failed to get attribute", "error", err2)
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if e2, err2 = h.cache.UpdateEntityStatus(uid, models.StatusFromString(e1.Status), user); err2 != nil {
		h.logger.Error("failed to save entity status", "error", err2)
		return h.error(req, fiber.StatusBadRequest, err2)
	}

	return h.respond(req, e2.ToOAS())
}

// PostEntityRestore implements the AttributesHandler interface.
func (h *entitiesHandler) PostEntityRestore(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	ns, id, err := h.checkUID(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	version, err := h.checkVersion(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	e, err2 := h.cache.RestoreEntityVersion(models.EntityUID(ns, id), version, user)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	return h.respond(req, e)
}

// DeleteEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) DeleteEntity(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	ns, id, err := h.checkUID(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	uid := models.EntityUID(ns, id)

	prev, _, err2 := h.cache.GetEntity(uid)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if prev == nil {
		if !req.QueryBool(ParamIgnoreMissing) {
			return h.error(req, fiber.StatusNotFound, errEntityNotFound)
		} else {
			return h.respond(req, models.NewEntity(ns, id, models.NewAttributeSet()).ToOAS())
		}
	}

	if prev, err2 = h.cache.RemoveEntity(uid, user); err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	return h.respond(req, prev.ToOAS())
}

func (h *entitiesHandler) checkUID(req *fiber.Ctx) (ns, id string, err error) {
	ns = req.Params("type")
	if ns == "" || len(ns) > 80 {
		return "", "", errEntityTypeError
	}

	id = req.Params("id")
	if id == "" || len(id) > 200 {
		return "", "", errEntityIDError
	}

	return ns, id, nil
}

func (h *entitiesHandler) checkVersion(req *fiber.Ctx) (int, error) {
	v, err := req.ParamsInt("version")
	if err != nil {
		return 0, err
	}

	if v <= 0 {
		return 0, errEntityVersionError
	}

	return v, nil
}

// checkBody parses and validates the request body, returning the Code and error to report via
// badRequest on failure. A nil error means the returned Entity is valid.
func (h *entitiesHandler) checkBody(req *fiber.Ctx, ns, id string) (*oas.Entity, string, error) {
	var e oas.Entity
	if err := req.BodyParser(&e); err != nil {
		return nil, codeBadRequest, err
	}

	chk := newFieldChecker().
		checkIdentifiers(ns, &e.Type, errEntityTypeMismatch).
		checkIdentifiers(id, &e.Id, errEntityIDMismatch).
		checkEntityType(e.Type).
		checkStatus(e.Status).
		checkTitle(e.Metadata.Title).
		checkTags(e.Metadata.Tags).
		checkAttributes(e.Attributes)

	if chk.checkFailed() {
		return nil, chk.firstCode(), chk.error()
	}

	return &e, "", nil
}

// checkBodyStatus is like checkBody but for the status-only request body.
func (h *entitiesHandler) checkBodyStatus(req *fiber.Ctx, ns, id string) (*oas.EntityStatus, string, error) {
	var e oas.EntityStatus
	if err := req.BodyParser(&e); err != nil {
		return nil, codeBadRequest, err
	}

	chk := newFieldChecker().
		checkIdentifiers(ns, &e.Type, errEntityTypeMismatch).
		checkIdentifiers(id, &e.Id, errEntityIDMismatch).
		checkStatus(e.Status)

	if chk.checkFailed() {
		return nil, chk.firstCode(), chk.error()
	}

	return &e, "", nil
}

func (h *entitiesHandler) authorize(req *fiber.Ctx) (identity.Principal, error) {
	return authorizeRequest(h.authorizer, req, h.logger)
}

func (h *entitiesHandler) error(req *fiber.Ctx, status int, err error) error {
	h.logger.Error("request error", "path", req.Path(), "err", err, "status", status)
	return server.SendMessageResponse(req, status, err.Error())
}

// badRequest logs a warning and returns a 400 with the given validation error's code and message.
func (h *entitiesHandler) badRequest(req *fiber.Ctx, code string, err error) error {
	h.logger.Warn("entity request rejected", "path", req.Path(), "err", err, "status", fiber.StatusBadRequest)
	return server.SendProblemResponse(req, fiber.StatusBadRequest, code, err.Error())
}
