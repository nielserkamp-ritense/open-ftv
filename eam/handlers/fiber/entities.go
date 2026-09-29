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
	logger  *slog.Logger
	svc     *management.EntityService
	secured bool // an authorizer decides; the routes must carry the Identify middleware.
	principalResolver
}

// NewEntitiesHandler instantiates an entity handler. Every operation is decided by the entity
// service through the authorizer; a nil authorizer permits everything. *pip.PIP is an EntityStore.
func NewEntitiesHandler(logger *slog.Logger, store management.EntityStore, authorizer authorization.RequestAuthorizer, opts ...HandlerOption) EntitiesHandler {
	return &entitiesHandler{
		logger:            logger,
		svc:               management.NewEntityService(logger, store, authorizer),
		secured:           authorizer != nil,
		principalResolver: newPrincipalResolver(logger, opts),
	}
}

// caller is the request-scoped Principal this request runs as.
func (h *entitiesHandler) caller(req *fiber.Ctx) *authorization.RequestPrincipal {
	return callerOf(req, h.secured, h.logger)
}

// GetEntities implements the EntitiesHandler interface.
func (h *entitiesHandler) GetEntities(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	list, err := h.svc.List(req.UserContext(), h.caller(req))
	if err != nil {
		return h.fail(req, err)
	}

	resp := make([]*oas.Entity, len(list))
	for i := range list {
		resp[i] = list[i].ToOAS()
	}

	return h.respond(req, &resp)
}

// GetEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) GetEntity(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	uid, err := h.checkUID(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	details, err := h.svc.Get(req.UserContext(), h.caller(req), uid)
	if err != nil {
		return h.fail(req, err)
	}

	out := details.Entity.ToOAS()
	out.AuditLog = details.AuditLog
	out.UsageData = details.UsageData

	return h.respond(req, out)
}

// GetEntityVersions implements the EntitiesHandler interface.
func (h *entitiesHandler) GetEntityVersions(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	uid, err := h.checkUID(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	list, err := h.svc.Versions(req.UserContext(), h.caller(req), uid)
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, list)
}

// GetEntityVersion implements the EntitiesHandler interface.
func (h *entitiesHandler) GetEntityVersion(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	uid, err := h.checkUID(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	version, err := h.checkVersion(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	e, err := h.svc.Version(req.UserContext(), h.caller(req), uid, version)
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, e)
}

// PostEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) PostEntity(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	ns, id, err := h.checkTypeAndID(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	e, code, err := h.checkBody(req, ns, id)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	out, err := h.svc.Create(req.UserContext(), h.caller(req), e, req.QueryBool(ParamForceUpsert))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req.Status(fiber.StatusCreated), out.ToOAS())
}

// PutEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) PutEntity(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	ns, id, err := h.checkTypeAndID(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	e, code, err := h.checkBody(req, ns, id)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	out, err := h.svc.Update(req.UserContext(), h.caller(req), e, req.QueryBool(ParamForceUpsert))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, out.ToOAS())
}

// PatchEntityStatus implements the EntitiesHandler interface.
func (h *entitiesHandler) PatchEntityStatus(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	ns, id, err := h.checkTypeAndID(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	e, code, err := h.checkBodyStatus(req, ns, id)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	out, err := h.svc.SetStatus(req.UserContext(), h.caller(req), models.EntityUID(ns, id), models.StatusFromString(e.Status))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, out.ToOAS())
}

// PostEntityRestore implements the EntitiesHandler interface.
func (h *entitiesHandler) PostEntityRestore(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	uid, err := h.checkUID(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	version, err := h.checkVersion(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	out, err := h.svc.Restore(req.UserContext(), h.caller(req), uid, version)
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, out)
}

// DeleteEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) DeleteEntity(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	uid, err := h.checkUID(req)
	if err != nil {
		return h.error(req, fiber.StatusBadRequest, err)
	}

	out, err := h.svc.Delete(req.UserContext(), h.caller(req), uid, req.QueryBool(ParamIgnoreMissing))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, out.ToOAS())
}

// checkUID validates the type and id in the path and returns them as one UID.
func (h *entitiesHandler) checkUID(req *fiber.Ctx) (string, error) {
	ns, id, err := h.checkTypeAndID(req)
	if err != nil {
		return "", err
	}

	return models.EntityUID(ns, id), nil
}

func (h *entitiesHandler) checkTypeAndID(req *fiber.Ctx) (ns, id string, err error) {
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
		requireStatus(e.Status)

	if chk.checkFailed() {
		return nil, chk.firstCode(), chk.error()
	}

	return &e, "", nil
}

// fail answers a failed service call: a denial is 403 regardless of whether the entity exists,
// an invalid status transition 400, a missing entity 404, an existing one 409, and anything
// else (including a permitted caller that could not be recorded) 500 without the cause.
func (h *entitiesHandler) fail(req *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, authorization.ErrForbidden):
		return auth.Forbidden(req, err, h.logger)
	case errors.Is(err, management.ErrInvalidTransition):
		return h.error(req, fiber.StatusBadRequest, err)
	case errors.Is(err, management.ErrNotFound):
		return h.error(req, fiber.StatusNotFound, errEntityNotFound)
	case errors.Is(err, management.ErrExists):
		return h.error(req, fiber.StatusConflict, errEntityExists)
	default:
		return internalError(req, h.logger, err)
	}
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
