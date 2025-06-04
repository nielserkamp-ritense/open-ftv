package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/models"
)

// TablesVersion is the full semantic API version for the table endpoints.
const TablesVersion = "1.0.0"

type TablesHandler interface {
	GetTables(req *fiber.Ctx) error
	GetTable(req *fiber.Ctx) error
	PutTable(req *fiber.Ctx) error
	PostTable(req *fiber.Ctx) error
	DeleteTable(req *fiber.Ctx) error
}

// NewTableHandler instantiates a table handler.
func NewTableHandler(s store.Storage, logger *slog.Logger) TablesHandler {
	return &tableHandler{logger: logger, s: s}
}

// GetTables implements the TablesHandler interface.
func (h *tableHandler) GetTables(req *fiber.Ctx) error {
	req.Set(HeaderVersion, TablesVersion)

	sources := h.s.GetDatasources()

	var list models.Rows
	for _, source := range sources {
		filter, matcher := buildFilters(req, source.Definition())
		for _, table := range source.Tables {
			if r := table.AsRow(); r.MatchPrimary(filter) {
				list = append(list, r.MatchFields(matcher))
			}
		}
	}

	if len(list) == 0 {
		return server.SendMessageResponse(req, fiber.StatusNotFound, "no matching tables found")
	}
	return buildContent(req, list)
}

// GetTable implements the TablesHandler interface.
func (h *tableHandler) GetTable(req *fiber.Ctx) error {
	req.Set(HeaderVersion, TablesVersion)

	key := req.Params("key")
	if key == "" || len(key) > 100 {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, "invalid table key")
	}

	t, err := h.s.GetTable(key)
	if err != nil {
		return server.SendMessageResponse(req, fiber.StatusNotFound, err.Error())
	}

	return buildContent(req, t)
}

// PutTable implements the TablesHandler interface.
func (h *tableHandler) PutTable(req *fiber.Ctx) error {
	req.Set(HeaderVersion, TablesVersion)

	// TODO: ...

	return req.SendStatus(fiber.StatusNotImplemented)
}

// PostTable implements the TablesHandler interface.
func (h *tableHandler) PostTable(req *fiber.Ctx) error {
	req.Set(HeaderVersion, TablesVersion)

	// TODO: ...

	return req.SendStatus(fiber.StatusNotImplemented)
}

// DeleteTable implements the TablesHandler interface.
func (h *tableHandler) DeleteTable(req *fiber.Ctx) error {
	req.Set(HeaderVersion, TablesVersion)

	// TODO: ...

	return req.SendStatus(fiber.StatusNotImplemented)
}

type tableHandler struct {
	logger *slog.Logger
	s      store.Storage
}

// const (
// 	tableNotFound   = "table not found"
// 	tableExists     = "table already exists"
// 	tableIDError    = "table ID must be filled and less or equal to 100 characters"
// 	tableIDMismatch = "table ID mismatch"
// )
