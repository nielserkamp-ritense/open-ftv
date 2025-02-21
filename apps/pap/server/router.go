package server

import (
	"context"

	"github.com/gofiber/fiber/v2"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers/fiber"
)

// initRoutes sets up the routing table for HTTP requests.
func (s *service) initRoutes(ctx context.Context, svc *fiber.App) {
	s.ctx = ctx

	s.auth = New(s.ctx, s.cfg, s.logger)
	if s.auth == nil {
		panic("failed to initialize authorization handler")
	}

	p, err := NewPAP(s.ctx, s.cfg, s.logger)
	if err != nil {
		panic("failed to initialize PAP handler")
	}
	s.pap = p

	s.initHealth(svc)

	// API v1.
	v1 := svc.Group("/v1")
	s.initPolicies(v1)
}

func (s *service) initHealth(svc *fiber.App) {
	svc.Get("/healthz", handle.HealthZ)
}

func (s *service) initPolicies(v1 fiber.Router) {
	policies := handle.NewPoliciesHandler(s.logger, s.pap)

	// policies.
	v1.Get(handle.PathPolicies, policies.GetPolicies)
	v1.Get(handle.PathPolicy, policies.GetPolicy)
	v1.Put(handle.PathPolicy, policies.PutPolicy)
	v1.Post(handle.PathPolicy, policies.PostPolicy)
	v1.Delete(handle.PathPolicy, policies.DeletePolicy)
}
