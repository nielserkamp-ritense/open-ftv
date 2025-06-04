package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store"
)

// TableDataVersion is the full semantic API version for the table endpoints.
const TableDataVersion = "1.0.0"

type TableDataHandler interface {
	GetRecords(req *fiber.Ctx) error
	GetRecord(req *fiber.Ctx) error
	PutRecord(req *fiber.Ctx) error
	PostRecord(req *fiber.Ctx) error
	DeleteRecord(req *fiber.Ctx) error
}

// NewTableDataHandler instantiates a table data handler.
func NewTableDataHandler(s store.Storage, logger *slog.Logger) TableDataHandler {
	return &dataHandler{logger: logger, s: s}
}

// GetRecords implements the TableDataHandler interface.
func (h *dataHandler) GetRecords(req *fiber.Ctx) error {
	req.Set(HeaderVersion, TableDataVersion)

	key := req.Params("key")
	if key == "" || len(key) > 100 {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, "invalid table key")
	}

	_, err := h.s.GetTable(key)
	if err != nil {
		return server.SendMessageResponse(req, fiber.StatusNotFound, err.Error())
	}

	filter, matcher := buildFilters(req, nil)
	list, err2 := h.s.Search(key, filter, matcher)
	if err2 != nil {
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, err2.Error())
	}
	if len(list) == 0 {
		return server.SendMessageResponse(req, fiber.StatusNotFound, "no matching records found")
	}

	return buildContent(req, list)
}

// GetRecord implements the TableDataHandler interface.
func (h *dataHandler) GetRecord(req *fiber.Ctx) error {
	req.Set(HeaderVersion, TableDataVersion)

	// TODO: ...

	return req.SendStatus(fiber.StatusNotImplemented)
}

// PutRecord implements the TableDataHandler interface.
func (h *dataHandler) PutRecord(req *fiber.Ctx) error {
	req.Set(HeaderVersion, TableDataVersion)

	// TODO: ...

	return req.SendStatus(fiber.StatusNotImplemented)
}

// PostRecord implements the TableDataHandler interface.
func (h *dataHandler) PostRecord(req *fiber.Ctx) error {
	req.Set(HeaderVersion, TableDataVersion)

	// TODO: ...

	return req.SendStatus(fiber.StatusNotImplemented)
}

// DeleteRecord implements the TableDataHandler interface.
func (h *dataHandler) DeleteRecord(req *fiber.Ctx) error {
	req.Set(HeaderVersion, TableDataVersion)

	// TODO: ...

	return req.SendStatus(fiber.StatusNotImplemented)
}

type dataHandler struct {
	logger *slog.Logger
	s      store.Storage
}

// const (
// 	recordNotFound   = "record not found"
// 	recordExists     = "record already exists"
// 	recordIDError    = "record key must be filled and less or equal to 100 characters"
// 	recordIDMismatch = "record key mismatch"
// )
