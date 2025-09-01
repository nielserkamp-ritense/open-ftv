package server

import (
	"context"

	"github.com/gofiber/fiber/v2"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
)

func (s *service) initRoutes(_ context.Context, svc *fiber.App) {
	svc.Get(handle.PathHealthZ, s.chk.HealthZ)
	svc.Get(handle.PathLiveZ, s.chk.LiveZ)
	svc.Get(handle.PathReadyZ, s.chk.ReadyZ)

	v1 := svc.Group(handle.PathV1)
	v1.Get("/ledenlijst", ledenlijst)
}
