package server

import (
	"github.com/gofiber/fiber/v2"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers/fiber"
)

// initRoutes sets up the routing table for HTTP requests.
func (s *service) initRoutes(svc *fiber.App) {
	auth := New(s.ctx, s.cfg, s.logger, s.logboek)
	if auth == nil {
		panic("failed to initialize authorization handler")
	}

	policies := handle.NewPoliciesHandler(s.logger, auth.Controller().PAP())
	if policies == nil {
		panic("failed to initialize policies handler")
	}

	attributes := handle.NewAttributesHandler(s.logger, auth.Controller())
	if attributes == nil {
		panic("failed to initialize attributes handler")
	}

	entities := handle.NewEntitiesHandler(s.logger, auth.Controller())
	if entities == nil {
		panic("failed to initialize entities handler")
	}

	// liveness & readiness.
	svc.Get("/healthz", handle.HealthZ)

	// API v1.
	v1 := svc.Group("/v1")

	// FSC authorization.
	v1.Post("/auth", auth.AuthFSC)

	// policies.
	v1.Get(handle.PathPolicies, policies.GetPolicies)
	v1.Get(handle.PathPolicy, policies.GetPolicy)
	v1.Put(handle.PathPolicy, policies.PutPolicy)
	v1.Post(handle.PathPolicy, policies.PostPolicy)
	v1.Delete(handle.PathPolicy, policies.DeletePolicy)

	// attributes.
	v1.Get(handle.PathAttributes, attributes.GetAttributes)
	v1.Get(handle.PathAttribute, attributes.GetAttribute)
	v1.Put(handle.PathAttribute, attributes.PutAttribute)
	v1.Post(handle.PathAttribute, attributes.PostAttribute)
	v1.Delete(handle.PathAttribute, attributes.DeleteAttribute)

	// entities.
	v1.Get(handle.PathEntities, entities.GetEntities)
	v1.Get(handle.PathEntity, entities.GetEntity)
	v1.Put(handle.PathEntity, entities.PutEntity)
	v1.Post(handle.PathEntity, entities.PostEntity)
	v1.Delete(handle.PathEntity, entities.DeleteEntity)

	// AuthZEN
	authZen := svc.Group("/authzen")
	authZenV1 := authZen.Group("/v1")
	authZenV1.Post("/evaluation", auth.AuthZEN)
}
