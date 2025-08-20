package fiber

import (
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
	// TODO: log request/response to audit log.
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
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, EntitiesVersion)

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var ns, id string
	if ns, id, ok, err = h.checkUID(req); !ok {
		return err
	}

	e, _, err2 := h.cache.GetEntity(models.EntityUID(ns, id))
	if err2 != nil {
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, err2.Error())
	}
	if e == nil {
		return server.SendMessageResponse(req, fiber.StatusNotFound, entityNotFound)
	}
	return req.JSON(e.ToOAS())
}

// PostEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) PostEntity(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
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
	if e1, ok, err = h.checkBody(req, ns, id); !ok {
		return err
	}

	e2 := models.EntityFromOAS(e1)

	prev, _, err2 := h.cache.GetEntity(e2.UID())
	if err2 != nil {
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, err2.Error())
	}
	if prev != nil && !req.QueryBool("forceUpsert") {
		return server.SendMessageResponse(req, fiber.StatusConflict, entityExists)
	}

	if e2, err2 = h.cache.AddEntityWithUser(e2, user); err2 != nil {
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, err2.Error())
	}
	return req.Status(fiber.StatusCreated).JSON(e2.ToOAS())
}

// PutEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) PutEntity(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
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
	if e1, ok, err = h.checkBody(req, ns, id); !ok {
		return err
	}

	e2 := models.EntityFromOAS(e1)

	prev, _, err2 := h.cache.GetEntity(e2.UID())
	if err2 != nil {
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, err2.Error())
	}

	if prev == nil && !req.QueryBool("forceUpsert") {
		return server.SendMessageResponse(req, fiber.StatusNotFound, entityNotFound)
	}

	if e2, err2 = h.cache.AddEntityWithUser(e2, user); err2 != nil {
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, err2.Error())
	}
	return req.JSON(e2.ToOAS())
}

// DeleteEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) DeleteEntity(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
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
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, err2.Error())
	}
	if prev == nil {
		if !req.QueryBool("ignoreMissing") {
			return server.SendMessageResponse(req, fiber.StatusNotFound, entityNotFound)
		} else {
			return req.JSON(models.NewEntity(ns, id, models.NewAttributeSet()).ToOAS())
		}
	}

	if prev, err2 = h.cache.RemoveEntity(uid, user); err2 != nil {
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, err2.Error())
	}
	return req.JSON(prev.ToOAS())
}

func (h *entitiesHandler) checkUID(req *fiber.Ctx) (string, string, bool, error) {
	ns := req.Params("type")
	if ns == "" || len(ns) > 500 {
		return "", "", false, server.SendMessageResponse(req, fiber.StatusBadRequest, entityTypeError)
	}

	id := req.Params("id")
	if id == "" || len(id) > 500 {
		return "", "", false, server.SendMessageResponse(req, fiber.StatusBadRequest, entityIDError)
	}

	return ns, id, true, nil
}

func (h *entitiesHandler) checkBody(req *fiber.Ctx, ns, id string) (*oas.Entity, bool, error) {
	var a oas.Entity
	if err := req.BodyParser(&a); err != nil {
		return nil, false, server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	if a.Type != ns {
		if a.Type != "" {
			return nil, false, server.SendMessageResponse(req, fiber.StatusBadRequest, entityTypeMismatch)
		}
		a.Type = ns
	}

	if a.Id != id {
		if a.Id != "" {
			return nil, false, server.SendMessageResponse(req, fiber.StatusBadRequest, entityIDMismatch)
		}
		a.Id = id
	}

	return &a, true, nil
}

func (h *entitiesHandler) authorize(req *fiber.Ctx) (string, bool, error) {
	if h.authorizer == nil {
		return auth.SystemUser, true, nil
	}

	resp, err := h.authorizer.Authorize(auth.FormatRequest(req))

	// TODO: log authorization decision to auth-decision log.

	return auth.Check(req, resp, err, h.logger)
}

type entitiesHandler struct {
	logger     *slog.Logger
	cache      *pip.PIP
	authorizer authorization.Authorizer
}

const (
	entityNotFound     = "entity not found"
	entityExists       = "entity already exists"
	entityTypeError    = "entity type must be filled and less or equal 500 characters"
	entityIDError      = "entity ID must be filled and less or equal 500 characters"
	entityTypeMismatch = "entity type mismatch"
	entityIDMismatch   = "entity ID mismatch"
)
