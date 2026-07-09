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

	if s.cfg.Persist.Type != "" {
		var err error
		if s.store, err = s.cfg.Persist.NewStore(ctx); err != nil {
			panic("failed to create persistence store: " + err.Error())
		}
	}

	s.pip = s.newPIP()
	s.pap = s.newPAP()

	s.initHealth(svc)

	// API v1.
	v1 := svc.Group("/v1")
	s.initAttributes(v1)
	s.initEntities(v1)
	s.initPolicies(v1)
	s.initODRL(v1)
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

func (s *service) initEntities(group fiber.Router) {
	entities := handle.NewEntitiesHandler(s.logger, s.pip, s.auth.Authorizer())

	// entities CRUD.
	group.Get(handle.PathEntities, entities.GetEntities)
	group.Get(handle.PathEntity, entities.GetEntity)
	group.Put(handle.PathEntity, entities.PutEntity)
	group.Post(handle.PathEntity, entities.PostEntity)
	group.Delete(handle.PathEntity, entities.DeleteEntity)
}

func (s *service) initPolicies(group fiber.Router) {
	policies := handle.NewPoliciesHandler(s.logger, s.pap, s.auth.Authorizer())

	// policies CRUD.
	group.Get(handle.PathPolicies, policies.GetPolicies)
	group.Get(handle.PathPolicy, policies.GetPolicy)
	group.Put(handle.PathPolicy, policies.PutPolicy)
	group.Post(handle.PathPolicy, policies.PostPolicy)
	group.Delete(handle.PathPolicy, policies.DeletePolicy)

	// Content-addressable policy resolution for ADL replay (adl.core.policies {key: hash}).
	group.Get("/policy-hash/:hash", s.GetPolicyByHash)
}
