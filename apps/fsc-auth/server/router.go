package server

import (
	"context"

	"github.com/gofiber/fiber/v2"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
)

// initRoutes sets up the routing table for HTTP requests.
func (s *service) initRoutes(ctx context.Context, svc *fiber.App) {
	s.ctx = ctx

	s.initHealth(svc)

	// API v1.
	v1 := svc.Group(handle.PathV1)

	// authorization.
	auth := New(s.ctx, s.cfg, s.logger)
	if auth == nil {
		panic("failed to initialize authorization handler")
	}

	s.initAuth(svc, v1, auth)
	s.initPolicies(v1, auth)
	s.initAttributes(v1, auth)
	s.initEntities(v1, auth)
}

func (s *service) initHealth(svc *fiber.App) {
	// liveness & readiness.
	svc.Get(handle.PathHealthZ, s.chk.HealthZ)
	svc.Get(handle.PathLiveZ, s.chk.LiveZ)
	svc.Get(handle.PathReadyZ, s.chk.ReadyZ)
}

func (s *service) initAuth(svc *fiber.App, v1 fiber.Router, auth AuthHandler) {
	// FSC authorization.
	// TODO: *DEPRECATED* remove in 2026
	v1.Post("/auth", auth.AuthFSC)

	// AuthZEN
	authZen := svc.Group(handle.PathAuthZEN)
	authZenV1 := authZen.Group(handle.PathV1)
	authZenV1.Post(handle.PathEvaluation, auth.AuthZEN)
}

func (s *service) initPolicies(v1 fiber.Router, auth AuthHandler) {
	policies := handle.NewPoliciesHandler(s.logger, auth.Controller().GetPAP(), nil)

	// policies.
	v1.Get(handle.PathPolicies, policies.GetPolicies)
	v1.Get(handle.PathPolicy, policies.GetPolicy)
	v1.Put(handle.PathPolicy, policies.PutPolicy)
	v1.Post(handle.PathPolicy, policies.PostPolicy)
	v1.Delete(handle.PathPolicy, policies.DeletePolicy)
}

func (s *service) initAttributes(v1 fiber.Router, auth AuthHandler) {
	attributes := handle.NewAttributesHandler(s.logger, auth.Controller().GetPIP(), nil)

	// attributes.
	v1.Get(handle.PathAttributes, attributes.GetAttributes)
	v1.Get(handle.PathAttribute, attributes.GetAttribute)
	v1.Put(handle.PathAttribute, attributes.PutAttribute)
	v1.Post(handle.PathAttribute, attributes.PostAttribute)
	v1.Delete(handle.PathAttribute, attributes.DeleteAttribute)
}

func (s *service) initEntities(v1 fiber.Router, auth AuthHandler) {
	entities := handle.NewEntitiesHandler(s.logger, auth.Controller().GetPIP(), nil)

	// entities.
	v1.Get(handle.PathEntities, entities.GetEntities)
	v1.Get(handle.PathEntity, entities.GetEntity)
	v1.Put(handle.PathEntity, entities.PutEntity)
	v1.Post(handle.PathEntity, entities.PostEntity)
	v1.Delete(handle.PathEntity, entities.DeleteEntity)
}
