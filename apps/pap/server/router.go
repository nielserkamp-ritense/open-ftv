package server

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// initMainRoutes sets up the routing table for HTTP requests.
func (s *Services) initMainRoutes(ctx context.Context, svc *fiber.App) {
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

	v1 := svc.Group(handle.PathV1)
	s.initLanguages(v1)
	s.initTags(v1)
	s.initPolicies(v1)
	s.initDeployments(v1)
}

func (s *Services) initLanguages(group fiber.Router) {
	apis := handle.NewLanguagesHandler(s.logger, s.pap, s.auth.Authorizer())

	// languages CRUD.
	group.Get(handle.PathLanguages, apis.GetLanguages)
}

func (s *Services) initTags(group fiber.Router) {
	if tags := s.cfg.Tags(); len(tags) > 0 {
		if err := s.pap.LoadTags(tags); err != nil {
			s.logger.Error("failed to load tags", "error", err.Error())
		} else {
			s.logger.Info("tags loaded successfully", "count", len(tags))
		}
	}

	apis := handle.NewTagsHandler(s.logger, s.pap, s.auth.Authorizer())

	// tags CRUD.
	group.Get(handle.PathTags, apis.GetTags).
		Get(handle.PathTag, apis.GetTag).
		Put(handle.PathTag, apis.PutTag).
		Post(handle.PathTag, apis.PostTag).
		Delete(handle.PathTag, apis.DeleteTag)
}

func (s *Services) initPolicies(group fiber.Router) {
	apis := handle.NewPoliciesHandler(s.logger, s.pap, s.auth.Authorizer())

	// policies CRUD.
	group.Get(handle.PathPolicies, apis.GetPolicies).
		Get(handle.PathPolicy, apis.GetPolicy).
		Put(handle.PathPolicy, apis.PutPolicy).
		Post(handle.PathPolicy, apis.PostPolicy).
		Delete(handle.PathPolicy, apis.DeletePolicy)
}

func (s *Services) initDeployments(group fiber.Router) {
	s.bundleManager = bundles.NewManager(
		s.ctx,
		s.logger,
		bundles.WithConfig(s.cfg.BundlePath, s.cfg.BundleRecurse),
		bundles.WithPolicyHandler(s.pap),
		bundles.MaxWorkers(s.cfg.Workers),
		bundles.WithStageDelay(s.cfg.StageDelay),
		bundles.BundleTimeout(s.cfg.BundleTimeout),
	)

	apis := handle.NewBundlesHandler(s.logger, s.pap, s.bundleManager, s.auth.Authorizer())

	// restart the last interrupted bundle deployment run if needed.
	time.AfterFunc(5*time.Second, func() { s.pap.RestartDeployment(s.bundleManager) })

	// deployments CRUD.
	group.Get(handle.PathStatuses, apis.GetStatuses).
		Get(handle.PathCompressionTypes, apis.GetCompressTypes).
		Get(handle.PathConfigs, apis.GetConfigs).
		Get(handle.PathDeployments, apis.GetDeployments).
		Get(handle.PathLastDeployment, apis.GetLastDeployment).
		Get(handle.PathDeploymentID, apis.GetDeployment).
		Post(handle.PathDeployment, apis.PostDeployment)
}

// initBundleRoutes sets up the routing table for bundle retrieval requests.
func (s *Services) initBundleRoutes(_ context.Context, svc *fiber.App) {
	apis := handle.NewBundlesHandler(s.logger, s.pap, s.bundleManager, s.auth.Authorizer())

	// bundle retrieval for PDPs.
	v1 := svc.Group(handle.PathV1)
	v1.Get(handle.PathBundle, apis.GetBundle)
}

// initHealthRoutes sets up the routing table for health requests.
func (s *Services) initHealthRoutes(_ context.Context, svc *fiber.App) {
	// liveness and readiness.
	svc.Get(handle.PathHealthZ, s.chk.HealthZ).
		Get(handle.PathLiveZ, s.chk.HealthZ).
		Get(handle.PathReadyZ, s.chk.HealthZ)
}
