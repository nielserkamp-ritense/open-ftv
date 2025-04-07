package server

import (
	"context"

	"github.com/gofiber/fiber/v2"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// initRoutes sets up the routing table for HTTP requests.
func (s *service) initRoutes(ctx context.Context, svc *fiber.App) {
	s.ctx = ctx
	s.l = models.LanguageFromString(s.cfg.PAP.Language)

	auth := s.newAuth()
	if auth == nil {
		panic("failed to initialize authorization handler")
	}

	var err error
	if s.pip, err = s.newPIP(); err != nil {
		panic("failed to initialize PIP handler: " + err.Error())
	}

	s.initHealth(svc)

	// API v1.
	v1 := svc.Group("/v1")
	s.initAttributes(v1, auth)
	s.initEntities(v1, auth)
}

func (s *service) initHealth(svc *fiber.App) {
	// liveness & readiness.
	svc.Get("/healthz", handle.HealthZ)
}

func (s *service) initAttributes(group fiber.Router, auth AuthHandler) {
	attributes := handle.NewAttributesHandler(s.logger, s.pip)

	// attributes CRUD.
	group.Get(handle.PathAttributes, attributes.GetAttributes)
	group.Get(handle.PathAttribute, attributes.GetAttribute)
	group.Put(handle.PathAttribute, attributes.PutAttribute)
	group.Post(handle.PathAttribute, attributes.PostAttribute)
	group.Delete(handle.PathAttribute, attributes.DeleteAttribute)
}

func (s *service) initEntities(group fiber.Router, auth AuthHandler) {
	entities := handle.NewEntitiesHandler(s.logger, s.pip)

	// entities CRUD.
	group.Get(handle.PathEntities, entities.GetEntities)
	group.Get(handle.PathEntity, entities.GetEntity)
	group.Put(handle.PathEntity, entities.PutEntity)
	group.Post(handle.PathEntity, entities.PostEntity)
	group.Delete(handle.PathEntity, entities.DeleteEntity)
}
