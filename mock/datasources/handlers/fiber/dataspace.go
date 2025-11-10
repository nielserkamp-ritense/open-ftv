package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store"
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

	reqCtx, err := buildRequestContext(req, "")
	if err != nil {
		return server.SendMessageResponse(req, fiber.StatusBadRequest, err.Error())
	}

	d := h.s.GetDataspace()
	if d == nil {
		return server.SendMessageResponse(req, fiber.StatusNotFound, "no dataspace found")
	}
	return buildContent(req, d.AsRow().MatchFields(reqCtx.Matcher), nil)
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
