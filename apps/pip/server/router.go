package server

import (
	"context"

	"github.com/gofiber/fiber/v2"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// initMainRoutes sets up the routing table for main requests.
func (s *Services) initMainRoutes(ctx context.Context, svc *fiber.App) {
	s.ctx = ctx
	s.l = models.LanguageFromString(s.cfg.PAP.Language)

	s.auth = s.newAuth()
	if s.auth == nil {
		panic("failed to initialize authorization handler")
	}

	var err error
	if s.pip, err = s.newPIP(); err != nil {
		panic("failed to initialize PIP handler: " + err.Error())
	}

	// API v1.
	v1 := svc.Group(handle.PathV1)
	s.initAttributes(v1)
	s.initEntities(v1)
}

func (s *Services) initAttributes(group fiber.Router) {
	attributes := handle.NewAttributesHandler(s.logger, s.pip, s.auth.Authorizer())

	// attributes CRUD.
	group.Get(handle.PathAttributes, attributes.GetAttributes).
		Get(handle.PathAttribute, attributes.GetAttribute).
		Put(handle.PathAttribute, attributes.PutAttribute).
		Post(handle.PathAttribute, attributes.PostAttribute).
		Delete(handle.PathAttribute, attributes.DeleteAttribute)
}

func (s *Services) initEntities(group fiber.Router) {
	entities := handle.NewEntitiesHandler(s.logger, s.pip, s.auth.Authorizer())

	// entities CRUD.
	group.Get(handle.PathEntities, entities.GetEntities).
		Get(handle.PathEntity, entities.GetEntity).
		Put(handle.PathEntity, entities.PutEntity).
		Post(handle.PathEntity, entities.PostEntity).
		Delete(handle.PathEntity, entities.DeleteEntity)
}

// initHealthRoutes sets up the routing table for health requests.
func (s *Services) initHealthRoutes(_ context.Context, svc *fiber.App) {
	// liveness & readiness.
	svc.Get(handle.PathHealthZ, s.chk.HealthZ).
		Get(handle.PathLiveZ, s.chk.LiveZ).
		Get(handle.PathReadyZ, s.chk.ReadyZ)
}
