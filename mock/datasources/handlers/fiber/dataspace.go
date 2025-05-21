package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store"
)

// DataspaceVersion is the full semantic API version for the dataspace endpoints.
const DataspaceVersion = "1.0.0"

type DataspaceHandler interface {
	GetDataspace(req *fiber.Ctx) error
	PutDataspace(req *fiber.Ctx) error
	PostDataspace(req *fiber.Ctx) error
	DeleteDataspace(req *fiber.Ctx) error
}

// NewDataspaceHandler instantiates a dataspace handler.
func NewDataspaceHandler(s store.Storage, logger *slog.Logger) DataspaceHandler {
	return &dataspaceHandler{logger: logger, s: s}
}

// GetDataspace implements the DataspacesHandler interface.
func (h *dataspaceHandler) GetDataspace(req *fiber.Ctx) error {
	req.Set(HeaderVersion, DataspaceVersion)

	d := h.s.GetDataspace()
	if d == nil {
		return server.SendMessageResponse(req, fiber.StatusNotFound, "no dataspace found")
	}

	filter := buildFilter(req)
	_, matcher := extractFieldMatcher(filter)

	return buildContent(req, d.AsRecord().MatchFields(matcher))
}

// PutDataspace implements the DataspacesHandler interface.
func (h *dataspaceHandler) PutDataspace(req *fiber.Ctx) error {
	req.Set(HeaderVersion, DataspaceVersion)

	// TODO: ...

	return req.SendStatus(fiber.StatusNotImplemented)
}

// PostDataspace implements the DataspacesHandler interface.
func (h *dataspaceHandler) PostDataspace(req *fiber.Ctx) error {
	req.Set(HeaderVersion, DataspaceVersion)

	// TODO: ...

	return req.SendStatus(fiber.StatusNotImplemented)
}

// DeleteDataspace implements the DataspacesHandler interface.
func (h *dataspaceHandler) DeleteDataspace(req *fiber.Ctx) error {
	req.Set(HeaderVersion, DataspaceVersion)

	// TODO: ...

	return req.SendStatus(fiber.StatusNotImplemented)
}

type dataspaceHandler struct {
	logger *slog.Logger
	s      store.Storage
}

// const (
// 	dataspaceNotFound   = "dataspace not found"
// 	dataspaceExists     = "dataspace already exists"
// 	dataspaceIDError    = "dataspace ID must be filled and less or equal to 100 characters"
// 	dataspaceIDMismatch = "dataspace ID mismatch"
// )
