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

	var err error
	if s.pap, err = s.newPAP(); err != nil {
		panic("failed to initialize PAP handler")
	}

	s.initHealth(svc)

	// API v1.
	v1 := svc.Group("/v1")
	s.initPolicies(v1)
}

func (s *service) initHealth(svc *fiber.App) {
	// liveness & readiness.
	svc.Get("/healthz", handle.HealthZ)
}

func (s *service) initPolicies(group fiber.Router) {
	policies := handle.NewPoliciesHandler(s.logger, s.pap, s.auth.Authorizer())

	// policies CRUD.
	group.Get(handle.PathPolicies, policies.GetPolicies)
	group.Get(handle.PathPolicy, policies.GetPolicy)
	group.Put(handle.PathPolicy, policies.PutPolicy)
	group.Post(handle.PathPolicy, policies.PostPolicy)
	group.Delete(handle.PathPolicy, policies.DeletePolicy)
}
