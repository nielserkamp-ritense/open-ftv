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

// AttributesVersion is the full semantic API version for the attribute endpoints.
const AttributesVersion = "1.6.0" // check against oas/attributes/openapi.yaml!

// AttributesHandler represents the interface for handling requests about attributes.
type AttributesHandler interface {
	GetAttributes(req *fiber.Ctx) error
	GetAttribute(req *fiber.Ctx) error
	PostAttribute(req *fiber.Ctx) error
	PutAttribute(req *fiber.Ctx) error
	DeleteAttribute(req *fiber.Ctx) error
}

// NewAttributesHandler instantiates a policy handler.
func NewAttributesHandler(logger *slog.Logger, pip *pip.PIP, authorizer authorization.Authorizer) AttributesHandler {
	return &attributesHandler{logger: logger, cache: pip, authorizer: authorizer}
}

// GetAttributes implements the AttributesHandler interface.
func (h *attributesHandler) GetAttributes(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, AttributesVersion)

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	resp := make([]*oas.Attribute, 0, 32)
	h.cache.IterateAttributes(func(attr *models.Attribute) {
		resp = append(resp, attr.ToOAS())
	})

	return req.JSON(resp)
}

// GetAttribute implements the AttributesHandler interface.
func (h *attributesHandler) GetAttribute(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, AttributesVersion)

	_, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var key string
	if key, ok, err = h.checkKey(req); !ok {
		return err
	}

	a, _, err2 := h.cache.GetAttribute(key)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if a == nil {
		return h.error(req, fiber.StatusNotFound, attrNotFound)
	}
	return req.JSON(a.ToOAS())
}

// PostAttribute implements the AttributesHandler interface.
func (h *attributesHandler) PostAttribute(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, AttributesVersion)

	user, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var key string
	if key, ok, err = h.checkKey(req); !ok {
		return err
	}

	var a *oas.Attribute
	if a, ok, err = h.checkBody(req, key); !ok || err != nil {
		return err
	}

	a2, _, err2 := h.cache.GetAttribute(a.Key)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if a2 != nil && !req.QueryBool("forceUpsert") {
		return h.error(req, fiber.StatusConflict, attrExists)
	}

	if a2, err2 = h.cache.AddAttributeFromOAS(a, user); err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}
	return req.Status(fiber.StatusCreated).JSON(a2.ToOAS())
}

// PutAttribute implements the AttributesHandler interface.
func (h *attributesHandler) PutAttribute(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, AttributesVersion)

	user, ok, err := h.authorize(req)
	if !ok {
		return err
	}

	var key string
	key, ok, err = h.checkKey(req)
	if !ok {
		return err
	}

	var a *oas.Attribute
	if a, ok, err = h.checkBody(req, key); !ok || err != nil {
		return err
	}

	a2, _, err2 := h.cache.GetAttribute(a.Key)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if a2 == nil && !req.QueryBool("forceUpsert") {
		return h.error(req, fiber.StatusNotFound, attrNotFound)
	}

	if a2, err2 = h.cache.AddAttributeFromOAS(a, user); err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}
	return req.JSON(a2.ToOAS())
}

// DeleteAttribute implements the AttributesHandler interface.
func (h *attributesHandler) DeleteAttribute(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, AttributesVersion)

	user, ok, err := h.authorize(req)
	if !ok || err != nil {
		return err
	}

	var key string
	if key, ok, err = h.checkKey(req); !ok {
		return err
	}

	a2, _, err2 := h.cache.GetAttribute(key)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	if a2 == nil && !req.QueryBool("ignoreMissing") {
		return h.error(req, fiber.StatusNotFound, attrNotFound)
	}

	if a2, err2 = h.cache.RemoveAttribute(key, user); err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}
	return req.JSON(a2.ToOAS())
}

func (h *attributesHandler) checkKey(req *fiber.Ctx) (string, bool, error) {
	key := req.Params("key")
	if key == "" || len(key) > 500 {
		return "", false, h.error(req, fiber.StatusBadRequest, attrKeyError)
	}
	return key, true, nil
}

func (h *attributesHandler) checkBody(req *fiber.Ctx, key string) (*oas.Attribute, bool, error) {
	var a oas.Attribute
	if err := req.BodyParser(&a); err != nil {
		return nil, false, h.error(req, fiber.StatusBadRequest, err)
	}

	if a.Key != key {
		if a.Key != "" {
			return nil, false, h.error(req, fiber.StatusBadRequest, attrKeyMismatch)
		}
		a.Key = key
	}

	return &a, true, nil
}

func (h *attributesHandler) authorize(req *fiber.Ctx) (string, bool, error) {
	if h.authorizer == nil {
		return auth.SystemUser, true, nil
	}

	resp, err := h.authorizer.Authorize(auth.FormatRequest(req))

	// TODO: log authorization decision to auth-decision log.

	return auth.Check(req, resp, err, h.logger)
}

func (h *attributesHandler) error(req *fiber.Ctx, status int, err error) error {
	h.logger.Error("request error", "path", req.Path(), "err", err, "status", status)
	return server.SendMessageResponse(req, status, err.Error())
}

type attributesHandler struct {
	logger     *slog.Logger
	cache      *pip.PIP
	authorizer authorization.Authorizer
}

var (
	attrNotFound    = errors.New("attribute not found")
	attrExists      = errors.New("attribute already exists")
	attrKeyError    = errors.New("attribute key must be filled and less or equal 500 characters")
	attrKeyMismatch = errors.New("attribute key mismatch")
)
