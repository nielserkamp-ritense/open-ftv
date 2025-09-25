package fiber

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

// EntitiesVersion is the full semantic API version for the entity endpoints.
const EntitiesVersion = AttributesVersion

// EntitiesHandler represents the interface for handling requests about entities.
type EntitiesHandler interface {
	GetEntities(req *fiber.Ctx) error
	GetEntity(req *fiber.Ctx) error
	PostEntity(req *fiber.Ctx) error
	PutEntity(req *fiber.Ctx) error
	DeleteEntity(req *fiber.Ctx) error
}

// NewEntitiesHandler instantiates a policy handler.
func NewEntitiesHandler(logger *slog.Logger, pip *pip.PIP, authorizer authorization.Authorizer) EntitiesHandler {
	return &entitiesHandler{logger: logger, cache: pip, authorizer: authorizer}
}

// GetEntities implements the EntitiesHandler interface.
func (h *entitiesHandler) GetEntities(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	resp := make([]*oas.Entity, 0, 32)
	h.cache.IterateEntities(func(e *models.Entity) {
		resp = append(resp, e.ToOAS())
	})

	return req.JSON(&resp)
}

// GetEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) GetEntity(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var ns, id string
	if ns, id, ok, err = h.checkUID(req); !ok {
		return err
	}

	uid := models.EntityUID(ns, id)

	e, _, err2 := h.cache.GetEntity(uid)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if e == nil {
		return h.error(req, fiber.StatusNotFound, entityNotFound)
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

	return req.JSON(out)
}

// PostEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) PostEntity(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	user, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var ns, id string
	if ns, id, ok, err = h.checkUID(req); !ok {
		return err
	}

	var e1 *oas.Entity
	if e1, ok, err = h.checkBody(req, ns, id); !ok || e1 == nil {
		return err
	}

	prev, _, err2 := h.cache.GetEntity(models.EntityUID(e1.Type, e1.Id))
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if prev != nil && !req.QueryBool(ParamForceUpsert) {
		return h.error(req, fiber.StatusConflict, entityExists)
	}

	e2, err3 := h.cache.AddEntityFromOAS(e1, user)
	if err3 != nil {
		return h.error(req, fiber.StatusInternalServerError, err3)
	}
	return req.Status(fiber.StatusCreated).JSON(e2.ToOAS())
}

// PutEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) PutEntity(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	user, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var ns, id string
	if ns, id, ok, err = h.checkUID(req); !ok {
		return err
	}

	var e1 *oas.Entity
	if e1, ok, err = h.checkBody(req, ns, id); !ok || e1 == nil {
		return err
	}

	prev, _, err2 := h.cache.GetEntity(models.EntityUID(e1.Type, e1.Id))
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if prev == nil && !req.QueryBool(ParamForceUpsert) {
		return h.error(req, fiber.StatusNotFound, entityNotFound)
	}

	e2, err3 := h.cache.AddEntityFromOAS(e1, user)
	if err3 != nil {
		return h.error(req, fiber.StatusInternalServerError, err3)
	}
	return req.JSON(e2.ToOAS())
}

// DeleteEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) DeleteEntity(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EntitiesVersion)

	user, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var ns, id string
	if ns, id, ok, err = h.checkUID(req); !ok {
		return err
	}

	uid := models.EntityUID(ns, id)

	prev, _, err2 := h.cache.GetEntity(uid)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if prev == nil {
		if !req.QueryBool(ParamIgnoreMissing) {
			return h.error(req, fiber.StatusNotFound, entityNotFound)
		} else {
			return req.JSON(models.NewEntity(ns, id, models.NewAttributeSet()).ToOAS())
		}
	}

	if prev, err2 = h.cache.RemoveEntity(uid, user); err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}
	return req.JSON(prev.ToOAS())
}

func (h *entitiesHandler) checkUID(req *fiber.Ctx) (string, string, bool, error) {
	ns := req.Params("type")
	if ns == "" || len(ns) > 80 {
		return "", "", false, h.error(req, fiber.StatusBadRequest, entityTypeError)
	}

	id := req.Params("id")
	if id == "" || len(id) > 200 {
		return "", "", false, h.error(req, fiber.StatusBadRequest, entityIDError)
	}

	return ns, id, true, nil
}

func (h *entitiesHandler) checkBody(req *fiber.Ctx, ns, id string) (*oas.Entity, bool, error) {
	var e oas.Entity
	if err := req.BodyParser(&e); err != nil {
		return nil, false, h.error(req, fiber.StatusBadRequest, err)
	}

	chk := newFieldChecker().
		checkIdentifiers(ns, &e.Type, "entity type mismatch").
		checkIdentifiers(id, &e.Id, "entity id mismatch").
		checkTitle(e.Metadata.Title).
		checkTags(e.Metadata.Tags).
		checkAttributes(e.Attributes)

	if chk.checkFailed() {
		return nil, false, h.error(req, fiber.StatusBadRequest, chk.error())
	}
	return &e, true, nil
}

func (h *entitiesHandler) authorize(req *fiber.Ctx) (string, bool, error) {
	if h.authorizer == nil {
		return auth.SystemUser, true, nil
	}

	resp, err := h.authorizer.Authorize(auth.FormatRequest(req))
	return auth.Check(req, resp, err, h.logger)
}

func (h *entitiesHandler) error(req *fiber.Ctx, status int, err error) error {
	h.logger.Error("request error", "path", req.Path(), "err", err, "status", status)
	return server.SendMessageResponse(req, status, err.Error())
}

type entitiesHandler struct {
	logger     *slog.Logger
	cache      *pip.PIP
	authorizer authorization.Authorizer
}

var (
	entityNotFound  = errors.New("entity not found")
	entityExists    = errors.New("entity already exists")
	entityTypeError = errors.New("entity type must be filled and less or equal 80 characters")
	entityIDError   = errors.New("entity ID must be filled and less or equal 200 characters")
)
