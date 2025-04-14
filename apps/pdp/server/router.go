package server

import (
	"context"

	"github.com/gofiber/fiber/v2"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// initRoutes sets up the routing table for HTTP requests.
func (s *service) initRoutes(ctx context.Context, svc *fiber.App) {
	s.ctx = ctx
	s.l = models.LanguageFromString(s.cfg.PAP.Language)

	s.auth = s.newAuth()
	if s.auth == nil {
		panic("failed to initialize authorization handler")
	}

	s.initHealth(svc)
	s.initAuth(svc)
}

func (s *service) initHealth(svc *fiber.App) {
	// liveness & readiness.
	svc.Get("/healthz", handle.HealthZ)
}

func (s *service) initAuth(svc *fiber.App) {
	// AuthZEN
	authZen := svc.Group("/authzen")
	authZenV1 := authZen.Group("/v1")
	authZenV1.Post("/evaluation", s.auth.AuthZEN)
}
