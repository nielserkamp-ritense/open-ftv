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

	// API v1.
	v1 := svc.Group("/v1")
	s.initPolicies(v1, auth)
}

func (s *service) initHealth(svc *fiber.App) {
	// liveness & readiness.
	svc.Get("/healthz", handle.HealthZ)
}

func (s *service) initPolicies(v1 fiber.Router, auth AuthHandler) {
	policies := handle.NewPoliciesHandler(s.logger, auth.Controller().PAP())

	// policies.
	v1.Get(handle.PathPolicies, policies.GetPolicies)
	v1.Get(handle.PathPolicy, policies.GetPolicy)
	v1.Put(handle.PathPolicy, policies.PutPolicy)
	v1.Post(handle.PathPolicy, policies.PostPolicy)
	v1.Delete(handle.PathPolicy, policies.DeletePolicy)
}
