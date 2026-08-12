package fiber

import (
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store/memory/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

const EndpointVersion = "1.0.0"

// EndpointHandler represents the interface for handling a single endpoint definition.
type EndpointHandler interface {
	Handle(req *fiber.Ctx) error
}

// NewEndpointHandler instantiates an endpoint handler.
func NewEndpointHandler(s store.Storage, logger *slog.Logger, def *schema.Endpoint) EndpointHandler {
	return &endpointHandler{logger: logger, s: s, def: def}
}

// Handle implements the EndpointHandler interface.
func (h *endpointHandler) Handle(req *fiber.Ctx) error {
	req.Set(HeaderVersion, EndpointVersion)

	switch h.def.Type {
	case enums.GetMethod:
		return h.handleGet(req)
	case enums.PostMethod:
		return h.handlePost(req)
	case enums.PutMethod:
		return h.handlePut(req)
	case enums.PatchMethod:
		return h.handlePatch(req)
	case enums.DeleteMethod:
		return h.handleDelete(req)
	}

	return server.SendMessageResponse(req, fiber.StatusNotImplemented, "not implemented")
}

func (h *endpointHandler) handleGet(req *fiber.Ctx) error {
	if len(h.def.Keys) > 0 && pathHasKeys(h.def.Path, h.def.Keys) {
		return h.handleGetByPK(req)
	}

	reqCtx, err := buildRequestContext(req, h.def.Table)
	if err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	if err = reqCtx.Filter.Prepare(h.def.GetDatasource(), h.def.Joins); err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	list, modified, err2 := h.s.GetEndpoint(h.def, reqCtx)
	if err2 != nil {
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, err2.Error())
	}

	return buildContent(req, list, modified)
}

// handleGetByPK handles a GET endpoint that identifies a single record via
// path-parameter keys (h.def.Keys), e.g. GET /aanvraag/:id.
func (h *endpointHandler) handleGetByPK(req *fiber.Ctx) error {
	pk := pkValues(req, h.def.Keys)

	reqCtx, err := buildRequestContext(req, h.def.Table)
	if err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	if err = reqCtx.Filter.Prepare(h.def.GetDatasource(), h.def.Joins); err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	row, modified, err := h.s.GetEndpointByPK(h.def, pk, reqCtx)
	if err != nil {
		return server.SendMessageResponse(req, fiber.StatusNotFound, err.Error())
	}

	return buildContent(req, row, modified)
}

func (h *endpointHandler) handlePost(req *fiber.Ctx) error {
	body, err := decodeBody(req)
	if err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	pk := pkValues(req, h.def.Keys)
	if err = h.checkPK(pk, body, true); err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	if _, _, err = h.s.SelectPK(h.def.Table, pk, nil); err == nil {
		return server.SendMessageResponse(req, fiber.StatusConflict, "primary key already exists")
	}

	record := &models.Row{Data: body}
	if err = h.s.CreateRecord(h.def.Table, record); err != nil {
		return server.SendMessageResponse(req, fiber.StatusConflict, err.Error())
	}

	return server.SendBasicResponse(req, fiber.StatusCreated)
}

func (h *endpointHandler) handlePut(req *fiber.Ctx) error {
	body, err := decodeBody(req)
	if err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	pk := pkValues(req, h.def.Keys)
	if err = h.checkPK(pk, body, true); err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	if _, _, err = h.s.SelectPK(h.def.Table, pk, nil); err != nil {
		return server.SendMessageResponse(req, fiber.StatusNotFound, err.Error())
	}

	record := &models.Row{Data: body}
	if err = h.s.UpdateRecord(h.def.Table, pk, record); err != nil {
		return server.SendMessageResponse(req, fiber.StatusConflict, err.Error())
	}

	return server.SendBasicResponse(req, fiber.StatusNoContent)
}

func (h *endpointHandler) handlePatch(req *fiber.Ctx) error {
	body, err := decodeBody(req)
	if err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	pk := pkValues(req, h.def.Keys)
	if err = h.checkPK(pk, body, false); err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	old, _, err2 := h.s.SelectPK(h.def.Table, pk, nil)
	if err2 != nil {
		return server.SendMessageResponse(req, fiber.StatusNotFound, err2.Error())
	}

	if old, err = h.patchRecord(old, body); err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	if err = h.s.UpdateRecord(h.def.Table, pk, old); err != nil {
		return server.SendMessageResponse(req, fiber.StatusConflict, err.Error())
	}

	return server.SendBasicResponse(req, fiber.StatusNoContent)
}

func (h *endpointHandler) checkPK(pk []any, body map[string]any, mustEqual bool) error {
	table := h.def.Primary()

	for i, key := range table.PrimaryKey.Fields {
		s1 := convert.AnyToString(pk[i])
		if s1 == "" || s1 == "0" {
			return fmt.Errorf("key field %s must be filled", key)
		}

		v2, ok := body[key]
		if mustEqual {
			if !ok || convert.AnyToString(v2) != s1 {
				return fmt.Errorf("key field %s mismatch with record", key)
			}
		} else {
			s2 := convert.AnyToString(v2)
			if ok && s2 != "" && s2 != s1 {
				return fmt.Errorf("key field %s mismatch with record", key)
			}
		}
	}

	return nil
}

func (h *endpointHandler) patchRecord(old *models.Row, body map[string]any) (*models.Row, error) {
	table := h.def.Primary()

	for k, v := range body {
		field := table.Field(k)
		if field == nil {
			return nil, fmt.Errorf("field %s not found", k)
		}
		old.Data[k] = v
	}

	return old, nil
}

func (h *endpointHandler) handleDelete(req *fiber.Ctx) error {
	if len(h.def.Keys) > 0 && !pathHasKeys(h.def.Path, h.def.Keys) {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, "endpoint path does not declare all configured keys")
	}

	pk := pkValues(req, h.def.Keys)

	_, _, err := h.s.SelectPK(h.def.Table, pk, nil)
	if err != nil {
		return server.SendMessageResponse(req, fiber.StatusNotFound, err.Error())
	}

	if err = h.s.DeleteRecord(h.def.Table, pk); err != nil {
		return server.SendMessageResponse(req, fiber.StatusConflict, err.Error())
	}

	return server.SendBasicResponse(req, fiber.StatusNoContent)
}

type endpointHandler struct {
	s      store.Storage
	logger *slog.Logger
	def    *schema.Endpoint
}
