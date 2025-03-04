package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	fiber2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/attributes"
)

// AttributesVersion is the full semantic API version for the attribute endpoints.
const AttributesVersion = "1.0.0"

// AttributesHandler represents the interface for handling requests about attributes.
type AttributesHandler interface {
	GetAttributes(req *fiber.Ctx) error
	GetAttribute(req *fiber.Ctx) error
	PutAttribute(req *fiber.Ctx) error
	PostAttribute(req *fiber.Ctx) error
	DeleteAttribute(req *fiber.Ctx) error
}

// NewAttributesHandler instantiates a policy handler.
func NewAttributesHandler(logger *slog.Logger, pip pip.PIP) AttributesHandler {
	return &attributesHandler{logger: logger, cache: pip}
}

// GetAttributes implements the AttributesHandler interface.
func (h *attributesHandler) GetAttributes(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	resp := make([]*attributes.Attribute, 0, 32)
	h.cache.IterateAttributes(func(attr models.Attribute) {
		resp = append(resp, handlers.AttributeToOAS(attr))
	})

	if len(resp) == 0 {
		return fiber2.SendMessageResponse(req, fiber.StatusNotFound, attrNotFound)
	}
	return req.JSON(resp)
}

// GetAttribute implements the AttributesHandler interface.
func (h *attributesHandler) GetAttribute(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	key, ok, err := h.checkKey(req)
	if !ok {
		return err
	}

	attr := h.cache.GetAttribute(key)
	if attr == nil {
		return fiber2.SendMessageResponse(req, fiber.StatusNotFound, attrNotFound)
	}
	return req.JSON(handlers.AttributeToOAS(attr))
}

// PutAttribute implements the AttributesHandler interface.
func (h *attributesHandler) PutAttribute(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

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
		return fiber2.SendMessageResponse(req, fiber.StatusConflict, attrExists)
	}

	a := handlers.AttributeFromOAS(p)
	h.cache.AddAttribute(a.Key(), a.Value())
	return req.JSON(&attributes.Attribute{Key: a.Key(), Value: a.Value()})
}

// PostAttribute implements the AttributesHandler interface.
func (h *attributesHandler) PostAttribute(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

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
		return fiber2.SendMessageResponse(req, fiber.StatusNotFound, attrNotFound)
	}

	a := handlers.AttributeFromOAS(p)
	h.cache.AddAttributeWithType(a.Key(), a.Value(), a.Type())
	return req.JSON(handlers.AttributeToOAS(a))
}

// DeleteAttribute implements the AttributesHandler interface.
func (h *attributesHandler) DeleteAttribute(req *fiber.Ctx) error {
	req.Set(HeaderVersion, AttributesVersion)

	key, ok, err := h.checkKey(req)
	if !ok {
		return err
	}

	value := h.cache.GetAttribute(key)
	if value == nil && !req.QueryBool("ignoreMissing") {
		return fiber2.SendMessageResponse(req, fiber.StatusNotFound, attrNotFound)
	}

	h.cache.RemoveAttribute(key)
	return req.JSON(handlers.AttributeToOAS(models.NewAttribute(key, value)))
}

func (h *attributesHandler) checkKey(req *fiber.Ctx) (string, bool, error) {
	key := req.Params("key")
	if key == "" || len(key) > 500 {
		return "", false, fiber2.SendMessageResponse(req, fiber.StatusBadRequest, attrKeyError)
	}
	return key, true, nil
}

func (h *attributesHandler) checkBody(req *fiber.Ctx, key string) (*attributes.Attribute, bool, error) {
	var a attributes.Attribute
	if err := req.BodyParser(&a); err != nil {
		return nil, false, fiber2.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	if a.Key != key {
		if a.Key != "" {
			return nil, false, fiber2.SendMessageResponse(req, fiber.StatusBadRequest, attrKeyMismatch)
		}
		a.Key = key
	}

	return &a, true, nil
}

type attributesHandler struct {
	logger *slog.Logger
	cache  pip.PIP
}

const (
	attrNotFound    = "attribute not found"
	attrExists      = "attribute already exists"
	attrKeyError    = "attribute key must be filled and less or equal 500 characters"
	attrKeyMismatch = "attribute key mismatch"
)
