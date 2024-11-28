package server

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/handlers"
)

// initRoutes sets up the routing table for HTTP requests.
func (s *service) initRoutes() {
	auth := handlers.AuthHandler(s.cfg, s.logger, s.ldv)
	if auth == nil {
		panic("failed to initialize authorization handler")
	}

	// liveness & readiness.
	s.svc.Get("/healthz", handlers.HealthZ)

	// API v1.
	v1 := s.svc.Group("/v1")
	v1.Post("/auth", auth)
}
