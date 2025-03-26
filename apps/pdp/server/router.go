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
	s.initAuth(svc, auth)
}

func (s *service) initHealth(svc *fiber.App) {
	// liveness & readiness.
	svc.Get("/healthz", handle.HealthZ)
}

func (s *service) initAuth(svc *fiber.App, auth AuthHandler) {
	// AuthZEN
	authZen := svc.Group("/authzen")
	authZenV1 := authZen.Group("/v1")
	authZenV1.Post("/evaluation", auth.AuthZEN)
}
