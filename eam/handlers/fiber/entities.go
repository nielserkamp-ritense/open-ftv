package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/authorization"
	authRequest "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/attributes"
)

// EntitiesVersion is the full semantic API version for the entity endpoints.
const EntitiesVersion = "1.0.0"

// EntitiesHandler represents the interface for handling requests about entities.
type EntitiesHandler interface {
	GetEntities(req *fiber.Ctx) error
	GetEntity(req *fiber.Ctx) error
	PutEntity(req *fiber.Ctx) error
	PostEntity(req *fiber.Ctx) error
	DeleteEntity(req *fiber.Ctx) error
}

// NewEntitiesHandler instantiates a policy handler.
func NewEntitiesHandler(logger *slog.Logger, pip pip.PIP, authorizer authorization.Authorizer) EntitiesHandler {
	return &entitiesHandler{logger: logger, cache: pip, authorizer: authorizer}
}

// GetEntities implements the EntitiesHandler interface.
func (h *entitiesHandler) GetEntities(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, EntitiesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	resp := make([]*attributes.Entity, 0, 32)
	h.cache.IterateEntities(func(e models.Entity) {
		resp = append(resp, handlers.EntityToOAS(e))
	})

	if len(resp) == 0 {
		return server.SendMessageResponse(req, fiber.StatusNotFound, "no entities found")
	}
	return req.JSON(&resp)
}

// GetEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) GetEntity(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, EntitiesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	ns, id, ok, err := h.checkUID(req)
	if !ok {
		return err
	}

	e := h.cache.GetEntity(models.EntityUID(ns, id))
	if e == nil {
		return server.SendMessageResponse(req, fiber.StatusNotFound, entityNotFound)
	}
	return req.JSON(handlers.EntityToOAS(e))
}

// PutEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) PutEntity(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, EntitiesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	ns, id, ok, err := h.checkUID(req)
	if !ok {
		return err
	}

	var e1 *attributes.Entity
	if e1, ok, err = h.checkBody(req, ns, id); !ok {
		return err
	}

	e2 := handlers.EntityFromOAS(e1, h.cache.NewAttributeSet())

	value := h.cache.GetEntity(e2.UID())
	if value != nil && !req.QueryBool("forceUpsert") {
		return server.SendMessageResponse(req, fiber.StatusConflict, entityExists)
	}

	h.cache.AddEntity(e2)
	return req.JSON(handlers.EntityToOAS(e2))
}

// PostEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) PostEntity(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, EntitiesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	ns, id, ok, err := h.checkUID(req)
	if !ok {
		return err
	}

	var e1 *attributes.Entity
	if e1, ok, err = h.checkBody(req, ns, id); !ok {
		return err
	}

	e2 := handlers.EntityFromOAS(e1, h.cache.NewAttributeSet())

	e3 := h.cache.GetEntity(e2.UID())
	if e3 == nil && !req.QueryBool("forceUpsert") {
		return server.SendMessageResponse(req, fiber.StatusNotFound, entityNotFound)
	}

	h.cache.AddEntity(e2)
	return req.JSON(handlers.EntityToOAS(e2))
}

// DeleteEntity implements the EntitiesHandler interface.
func (h *entitiesHandler) DeleteEntity(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, EntitiesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	ns, id, ok, err := h.checkUID(req)
	if !ok {
		return err
	}

	uid := models.EntityUID(ns, id)

	e := h.cache.GetEntity(uid)
	if e == nil && !req.QueryBool("ignoreMissing") {
		return server.SendMessageResponse(req, fiber.StatusNotFound, entityNotFound)
	}

	if e == nil {
		e = models.NewEntity(ns, id, h.cache.NewAttributeSet())
	}

	h.cache.RemoveEntity(uid)
	return req.JSON(handlers.EntityToOAS(e))
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

func (h *entitiesHandler) checkBody(req *fiber.Ctx, ns, id string) (*attributes.Entity, bool, error) {
	var a attributes.Entity
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

func (h *entitiesHandler) authorize(req *fiber.Ctx) (bool, error) {
	if h.authorizer == nil {
		return true, nil
	}

	resp, err := h.authorizer.Authorize(authRequest.FormatRequest(req))

	// TODO: log authorization decision to audit log.

	return authRequest.Check(req, resp, err)
}

type entitiesHandler struct {
	logger     *slog.Logger
	cache      pip.PIP
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
