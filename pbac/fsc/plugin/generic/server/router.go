package server

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/handlers"
)

// initRoutes sets up the routing table for HTTP requests.
func (s *service) initRoutes() {
	auth := handlers.New(s.cfg, s.logger, s.logboek)
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

	// authorization.
	v1.Post("/auth", auth.AuthFSC)
	v1.Post("/authzen", auth.AuthZEN)

	// policies.
	v1.Get("/policies", policies.Policies)
}
