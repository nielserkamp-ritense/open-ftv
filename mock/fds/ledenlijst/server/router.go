package server

import (
	"github.com/gofiber/fiber/v2"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers/fiber"
)

func (s *service) initRoutes(svc *fiber.App) {
	svc.Get("/healthz", handle.HealthZ)

	v1 := svc.Group("/v1")
	v1.Get("/ledenlijst", ledenlijst)
}
