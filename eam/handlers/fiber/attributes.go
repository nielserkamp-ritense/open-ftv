package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	authRequest "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

// AttributesVersion is the full semantic API version for the attribute endpoints.
const AttributesVersion = "1.1.0" // check against oas/attributes/openapi.yaml!

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

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	resp := make([]*attributes.Attribute, 0, 32)
	h.cache.IterateAttributes(func(attr *models.Attribute) {
		resp = append(resp, handlers.AttributeToOAS(attr))
	})

	return req.JSON(resp)
}

// GetAttribute implements the AttributesHandler interface.
func (h *attributesHandler) GetAttribute(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, AttributesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	key, ok, err := h.checkKey(req)
	if !ok {
		return err
	}

	attr := h.cache.GetAttribute(key)
	if attr == nil {
		return server.SendMessageResponse(req, fiber.StatusNotFound, attrNotFound)
	}
	return req.JSON(handlers.AttributeToOAS(attr))
}

// PostAttribute implements the AttributesHandler interface.
func (h *attributesHandler) PostAttribute(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, AttributesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	key, ok, err := h.checkKey(req)
	if !ok {
		return err
	}

	var p *attributes.Attribute
	if p, ok, err = h.checkBody(req, key); !ok || err != nil {
		return err
	}

	value := h.cache.GetAttribute(p.Key)
	if value != nil && !req.QueryBool("forceUpsert") {
		return server.SendMessageResponse(req, fiber.StatusConflict, attrExists)
	}

	a := handlers.AttributeFromOAS(p)
	h.cache.AddAttribute(a.Key(), a.Value())
	return req.Status(fiber.StatusCreated).JSON(&attributes.Attribute{Key: a.Key(), Value: a.Value()})
}

// PutAttribute implements the AttributesHandler interface.
func (h *attributesHandler) PutAttribute(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, AttributesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	key, ok, err := h.checkKey(req)
	if !ok {
		return err
	}

	var p *attributes.Attribute
	if p, ok, err = h.checkBody(req, key); !ok || err != nil {
		return err
	}

	value := h.cache.GetAttribute(p.Key)
	if value == nil && !req.QueryBool("forceUpsert") {
		return server.SendMessageResponse(req, fiber.StatusNotFound, attrNotFound)
	}

	a := handlers.AttributeFromOAS(p)
	h.cache.AddAttributeWithType(a.Key(), a.Value(), a.Type())
	return req.JSON(handlers.AttributeToOAS(a))
}

// DeleteAttribute implements the AttributesHandler interface.
func (h *attributesHandler) DeleteAttribute(req *fiber.Ctx) error {
	// TODO: log request/response to audit log.
	req.Set(HeaderVersion, AttributesVersion)

	if ok, err := h.authorize(req); !ok || err != nil {
		return err
	}

	key, ok, err := h.checkKey(req)
	if !ok {
		return err
	}

	value := h.cache.GetAttribute(key)
	if value == nil && !req.QueryBool("ignoreMissing") {
		return server.SendMessageResponse(req, fiber.StatusNotFound, attrNotFound)
	}

	h.cache.RemoveAttribute(key)
	return req.JSON(handlers.AttributeToOAS(models.NewAttribute(key, value)))
}

func (h *attributesHandler) checkKey(req *fiber.Ctx) (string, bool, error) {
	key := req.Params("key")
	if key == "" || len(key) > 500 {
		return "", false, server.SendMessageResponse(req, fiber.StatusBadRequest, attrKeyError)
	}
	return key, true, nil
}

func (h *attributesHandler) checkBody(req *fiber.Ctx, key string) (*attributes.Attribute, bool, error) {
	var a attributes.Attribute
	if err := req.BodyParser(&a); err != nil {
		return nil, false, server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	if a.Key != key {
		if a.Key != "" {
			return nil, false, server.SendMessageResponse(req, fiber.StatusBadRequest, attrKeyMismatch)
		}
		a.Key = key
	}

	return &a, true, nil
}

func (h *attributesHandler) authorize(req *fiber.Ctx) (bool, error) {
	if h.authorizer == nil {
		return true, nil
	}

	resp, err := h.authorizer.Authorize(authRequest.FormatRequest(req))

	// TODO: log authorization decision to auth-decision log.

	return authRequest.Check(req, resp, err)
}

type attributesHandler struct {
	logger     *slog.Logger
	cache      *pip.PIP
	authorizer authorization.Authorizer
}

const (
	attrNotFound    = "attribute not found"
	attrExists      = "attribute already exists"
	attrKeyError    = "attribute key must be filled and less or equal 500 characters"
	attrKeyMismatch = "attribute key mismatch"
)
