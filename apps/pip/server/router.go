package server

import (
	"context"

	"github.com/gofiber/fiber/v2"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// initRoutes sets up the routing table for HTTP requests.
func (s *service) initRoutes(ctx context.Context, svc *fiber.App) {
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

	s.initHealth(svc)

	// API v1.
	v1 := svc.Group("/v1")
	s.initAttributes(v1)
	s.initEntities(v1)
	s.initEvents(v1)
}

func (s *service) initHealth(svc *fiber.App) {
	// liveness & readiness.
	svc.Get("/healthz", handle.HealthZ)
}

func (s *service) initAttributes(group fiber.Router) {
	attributes := handle.NewAttributesHandler(s.logger, s.pip, s.auth.Authorizer())

	// attributes CRUD.
	group.Get(handle.PathAttributes, attributes.GetAttributes)
	group.Get(handle.PathAttribute, attributes.GetAttribute)
	group.Put(handle.PathAttribute, attributes.PutAttribute)
	group.Post(handle.PathAttribute, attributes.PostAttribute)
	group.Delete(handle.PathAttribute, attributes.DeleteAttribute)
}

func (s *service) initEvents(group fiber.Router) {
	events := newEventsHandler(s.logger, s.pip, s.auth.Authorizer(), s.warc)

	// CloudEvents push-ingest & source-reference lookup.
	group.Post("/events", events.PostEvent)
	group.Get("/sourcerefs", events.GetSourceRefs)
	group.Get("/sourceref", events.GetSourceRef)
}

func (s *service) initEntities(group fiber.Router) {
	entities := handle.NewEntitiesHandler(s.logger, s.pip, s.auth.Authorizer())

	// entities CRUD.
	group.Get(handle.PathEntities, entities.GetEntities)
	group.Get(handle.PathEntity, entities.GetEntity)
	group.Put(handle.PathEntity, entities.PutEntity)
	group.Post(handle.PathEntity, entities.PostEntity)
	group.Delete(handle.PathEntity, entities.DeleteEntity)
}
