package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/attributes"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/handlers"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
)

// AttributesHandler represents the interface for handling requests about attributes.
type AttributesHandler interface {
	GetAttributes(req *fiber.Ctx) error
	GetAttribute(req *fiber.Ctx) error
	PutAttribute(req *fiber.Ctx) error
	PostAttribute(req *fiber.Ctx) error
	DeleteAttribute(req *fiber.Ctx) error
}

// NewAttributesHandler instantiates a policy handler.
func NewAttributesHandler(logger *slog.Logger, controller pdp.Controller) AttributesHandler {
	return &attributesHandler{logger: logger, controller: controller, cache: controller.PIP()}
}

// GetAttributes implements the AttributesHandler interface.
func (h *attributesHandler) GetAttributes(req *fiber.Ctx) error {
	resp := make([]*attributes.Attribute, 0, 32)
	h.cache.IterateAttributes(func(key string, value any) {
		resp = append(resp, handlers.AttributeToOAS(&models.Attribute{Key: key, Value: value}))
	})

	if len(resp) == 0 {
		return SendMessageResponse(req, fiber.StatusNotFound, attrNotFound)
	}
	return req.JSON(resp)
}

// GetAttribute implements the AttributesHandler interface.
func (h *attributesHandler) GetAttribute(req *fiber.Ctx) error {
	key, ok, err := h.checkKey(req)
	if !ok {
		return err
	}

	value := h.cache.GetAttribute(key)
	if value == nil {
		return SendMessageResponse(req, fiber.StatusNotFound, attrNotFound)
	}
	return req.JSON(handlers.AttributeToOAS(&models.Attribute{Key: key, Value: value}))
}

// PutAttribute implements the AttributesHandler interface.
func (h *attributesHandler) PutAttribute(req *fiber.Ctx) error {
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
		return SendMessageResponse(req, fiber.StatusConflict, attrExists)
	}

	a := handlers.AttributeFromOAS(p)
	h.cache.AddAttribute(a.Key, a.Value)
	return req.JSON(&attributes.Attribute{Key: a.Key, Value: a.Value})
}

// PostAttribute implements the AttributesHandler interface.
func (h *attributesHandler) PostAttribute(req *fiber.Ctx) error {
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
		return SendMessageResponse(req, fiber.StatusNotFound, attrNotFound)
	}

	a := handlers.AttributeFromOAS(p)
	h.cache.AddAttribute(a.Key, a.Value)
	return req.JSON(handlers.AttributeToOAS(a))
}

// DeleteAttribute implements the AttributesHandler interface.
func (h *attributesHandler) DeleteAttribute(req *fiber.Ctx) error {
	key, ok, err := h.checkKey(req)
	if !ok {
		return err
	}

	value := h.cache.GetAttribute(key)
	if value == nil && !req.QueryBool("ignoreMissing") {
		return SendMessageResponse(req, fiber.StatusNotFound, attrNotFound)
	}

	h.cache.RemoveAttribute(key)
	return req.JSON(handlers.AttributeToOAS(&models.Attribute{Key: key, Value: value}))
}

func (h *attributesHandler) checkKey(req *fiber.Ctx) (string, bool, error) {
	key := req.Params("key")
	if key == "" || len(key) > 500 {
		return "", false, SendMessageResponse(req, fiber.StatusBadRequest, attrKeyError)
	}
	return key, true, nil
}

func (h *attributesHandler) checkBody(req *fiber.Ctx, key string) (*attributes.Attribute, bool, error) {
	var a attributes.Attribute
	if err := req.BodyParser(&a); err != nil {
		return nil, false, SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	if a.Key != key {
		if a.Key != "" {
			return nil, false, SendMessageResponse(req, fiber.StatusBadRequest, attrKeyMismatch)
		}
		a.Key = key
	}

	return &a, true, nil
}

type attributesHandler struct {
	logger     *slog.Logger
	controller pdp.Controller
	cache      pip.PIP
}

const (
	attrNotFound    = "attribute not found"
	attrExists      = "attribute already exists"
	attrKeyError    = "attribute key must be filled and less or equal 500 characters"
	attrKeyMismatch = "attribute key mismatch"
)
