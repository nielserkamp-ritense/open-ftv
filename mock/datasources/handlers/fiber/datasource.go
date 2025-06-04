package fiber

import (
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/models"
)

// DatasourcesVersion is the full semantic API version for the datasource endpoints.
const DatasourcesVersion = "1.0.0"

type DatasourcesHandler interface {
	GetDatasources(req *fiber.Ctx) error
	GetDatasource(req *fiber.Ctx) error
	PutDatasource(req *fiber.Ctx) error
	PostDatasource(req *fiber.Ctx) error
	DeleteDatasource(req *fiber.Ctx) error
}

// NewDatasourceHandler instantiates a datasource handler.
func NewDatasourceHandler(s store.Storage, logger *slog.Logger) DatasourcesHandler {
	return &datasourceHandler{logger: logger, s: s}
}

// GetDatasources implements the DatasourcesHandler interface.
func (h *datasourceHandler) GetDatasources(req *fiber.Ctx) error {
	req.Set(HeaderVersion, DatasourcesVersion)

	sources := h.s.GetDatasources()

	var list models.Rows
	for _, source := range sources {
		filter, matcher := buildFilters(req, source.Definition())
		if r := source.AsRow(); r.MatchPrimary(filter) {
			list = append(list, r.MatchFields(matcher))
		}
	}

	if len(list) == 0 {
		return server.SendMessageResponse(req, fiber.StatusNotFound, "no matching datasources found")
	}
	return buildContent(req, list)
}

// GetDatasource implements the DatasourcesHandler interface.
func (h *datasourceHandler) GetDatasource(req *fiber.Ctx) error {
	req.Set(HeaderVersion, DatasourcesVersion)

	key := req.Params("key")
	if key == "" || len(key) > 100 {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, "invalid datasource key")
	}

	d := h.s.GetDatasource(key)
	if d == nil {
		return server.SendMessageResponse(req, fiber.StatusNotFound, fmt.Sprintf("datasource [%s] not found", key))
	}

	return buildContent(req, d)
}

// PutDatasource implements the DatasourcesHandler interface.
func (h *datasourceHandler) PutDatasource(req *fiber.Ctx) error {
	req.Set(HeaderVersion, DatasourcesVersion)

	// TODO: ...

	return req.SendStatus(fiber.StatusNotImplemented)
}

// PostDatasource implements the DatasourcesHandler interface.
func (h *datasourceHandler) PostDatasource(req *fiber.Ctx) error {
	req.Set(HeaderVersion, DatasourcesVersion)

	// TODO: ...

	return req.SendStatus(fiber.StatusNotImplemented)
}

// DeleteDatasource implements the DatasourcesHandler interface.
func (h *datasourceHandler) DeleteDatasource(req *fiber.Ctx) error {
	req.Set(HeaderVersion, DatasourcesVersion)

	// TODO: ...

	return req.SendStatus(fiber.StatusNotImplemented)
}

type datasourceHandler struct {
	logger *slog.Logger
	s      store.Storage
}

// const (
// 	datasourceNotFound   = "datasource not found"
// 	datasourceExists     = "datasource already exists"
// 	datasourceIDError    = "datasource ID must be filled and less or equal to 100 characters"
// 	datasourceIDMismatch = "datasource ID mismatch"
// )
