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

	v1 := svc.Group("/v1")

	s.initHealth(svc)
	s.initAuth(svc)
	s.initBundles(v1)
}

func (s *service) initHealth(svc *fiber.App) {
	// liveness and readiness.
	svc.Get("/healthz", handle.HealthZ)
}

func (s *service) initAuth(svc *fiber.App) {
	// AuthZEN
	authZen := svc.Group("/authzen")
	authZenV1 := authZen.Group("/v1")
	authZenV1.Post("/evaluation", s.auth.AuthZEN)
}

func (s *service) initBundles(v1 fiber.Router) {
	handler := handle.NewBundleReceiverHandler(s.logger, s.auth.Controller(), s.auth.Authorizer())
	v1.Post("/bundle", handler.PostBundle)
}
