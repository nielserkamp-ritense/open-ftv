package fiber

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store"
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

	list, err2 := h.s.GetEndpoint(h.def, buildFilter(req))
	if err2 != nil {
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, err2.Error())
	}
	if len(list) == 0 {
		return server.SendMessageResponse(req, fiber.StatusNotFound, "no matching records found")
	}

	return buildContent(req, list)
}

type endpointHandler struct {
	s      store.Storage
	logger *slog.Logger
	def    *schema.Endpoint
}
