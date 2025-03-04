package server

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities-no-ci/opensearch"
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
	s.initAttributes(v1, auth)
	s.initEntities(v1, auth)

	if s.cfg.OpenSearchIndex != "" {
		s.initAuthlog(v1)
	}
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

func (s *service) initAttributes(v1 fiber.Router, auth AuthHandler) {
	attributes := handle.NewAttributesHandler(s.logger, auth.Controller().PIP())

	// attributes.
	v1.Get(handle.PathAttributes, attributes.GetAttributes)
	v1.Get(handle.PathAttribute, attributes.GetAttribute)
	v1.Put(handle.PathAttribute, attributes.PutAttribute)
	v1.Post(handle.PathAttribute, attributes.PostAttribute)
	v1.Delete(handle.PathAttribute, attributes.DeleteAttribute)
}

func (s *service) initEntities(v1 fiber.Router, auth AuthHandler) {
	entities := handle.NewEntitiesHandler(s.logger, auth.Controller().PIP())

	// entities.
	v1.Get(handle.PathEntities, entities.GetEntities)
	v1.Get(handle.PathEntity, entities.GetEntity)
	v1.Put(handle.PathEntity, entities.PutEntity)
	v1.Post(handle.PathEntity, entities.PostEntity)
	v1.Delete(handle.PathEntity, entities.DeleteEntity)
}

func (s *service) initAuthlog(v1 fiber.Router) {
	searcher, err := opensearch.NewSearcher(s.cfg.OpenSearchUser, s.cfg.OpenSearchPswd, strings.Split(s.cfg.OpenSearchEndpoints, ","))
	if err != nil {
		s.logger.Error("failed to initialize OpenSearch", "user", s.cfg.OpenSearchUser, "endpoints", s.cfg.OpenSearchEndpoints, "error", err)
		return
	}

	authlog := handle.NewAuthlogHandler(s.logger, s.cfg.OpenSearchIndex, searcher)

	// authlog
	auth := v1.Group(handle.PathAuthlog)
	auth.Get(handle.PathResource, authlog.GetAuthlogResource)
}
