package server

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/handlers"
)

// initRoutes sets up the routing table for HTTP requests.
func (s *service) initRoutes() {
	auth := handlers.New(s.ctx, s.cfg, s.logger, s.logboek)
	if auth == nil {
		panic("failed to initialize authorization handler")
	}

	policies := handlers.NewPoliciesHandler(s.cfg, s.logger, auth.Controller())
	if policies == nil {
		panic("failed to initialize policies handler")
	}

	// liveness & readiness.
	s.svc.Get("/healthz", handlers.HealthZ)

	// API v1.
	v1 := s.svc.Group("/v1")

	// FSC authorization.
	v1.Post("/auth", auth.AuthFSC)

	// policies.
	v1.Get("/policies", policies.GetPolicies)
	v1.Get("/policy/:id", policies.GetPolicy)
	v1.Put("/policy/:id", policies.PutPolicy)
	v1.Post("/policy/:id", policies.PostPolicy)
	v1.Delete("/policy/:id", policies.DeletePolicy)

	// AuthZEN
	authZen := s.svc.Group("/authzen")
	authZenV1 := authZen.Group("/v1")
	authZenV1.Post("/evaluation", auth.AuthZEN)
}
