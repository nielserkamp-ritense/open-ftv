package server

import (
	"context"

	"github.com/gofiber/fiber/v2"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers/fiber"
)

// initRoutes sets up the routing table for HTTP requests.
func (s *service) initRoutes(ctx context.Context, svc *fiber.App) {
	s.ctx = ctx

	s.initHealth(svc)

	auth := New(s.ctx, s.cfg, s.logger)
	if auth == nil {
		panic("failed to initialize authorization handler")
	}

	p, err := NewPIP(s.ctx, s.cfg, s.logger)
	if err != nil {
		panic("failed to initialize PIP handler")
	}
	s.pip = p

	// API v1.
	v1 := svc.Group("/v1")
	s.initAttributes(v1, auth)
	s.initEntities(v1, auth)
}

func (s *service) initHealth(svc *fiber.App) {
	// liveness & readiness.
	svc.Get("/healthz", handle.HealthZ)
}

func (s *service) initAttributes(v1 fiber.Router, auth AuthHandler) {
	attributes := handle.NewAttributesHandler(s.logger, s.pip)

	// attributes.
	v1.Get(handle.PathAttributes, attributes.GetAttributes)
	v1.Get(handle.PathAttribute, attributes.GetAttribute)
	v1.Put(handle.PathAttribute, attributes.PutAttribute)
	v1.Post(handle.PathAttribute, attributes.PostAttribute)
	v1.Delete(handle.PathAttribute, attributes.DeleteAttribute)
}

func (s *service) initEntities(v1 fiber.Router, auth AuthHandler) {
	entities := handle.NewEntitiesHandler(s.logger, s.pip)

	// entities.
	v1.Get(handle.PathEntities, entities.GetEntities)
	v1.Get(handle.PathEntity, entities.GetEntity)
	v1.Put(handle.PathEntity, entities.PutEntity)
	v1.Post(handle.PathEntity, entities.PostEntity)
	v1.Delete(handle.PathEntity, entities.DeleteEntity)
}
