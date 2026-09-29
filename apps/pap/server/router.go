package server

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	authfiber "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
)

// initMainRoutes sets up the routing table for HTTP requests.
func (s *Services) initMainRoutes(ctx context.Context, svc *fiber.App) {
	s.ctx = ctx
	s.l = models.LanguageFromString(s.cfg.Language)

	var err error
	if s.pap, err = s.newPAP(); err != nil {
		s.logger.Error("failed to initialize PAP handler", "error", err)
		panic("failed to initialize PAP handler: " + err.Error())
	}

	s.auth = s.newAuth()
	if s.auth == nil {
		s.logger.Error("failed to initialize authorization handler")
		panic("failed to initialize authorization handler")
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
	// A missing or unreadable tags path is not fatal: whatever was read before the error is still loaded.
	tags, err := pap.ReadTagFiles(s.cfg.TagsPath)
	if err != nil {
		s.logger.Warn("failed to read tags", "path", s.cfg.TagsPath, "error", err.Error())
	}

	if len(tags) > 0 {
		if err = s.pap.LoadTags(tags); err != nil {
			s.logger.Error("failed to load tags", "error", err.Error())
		} else {
			s.logger.Info("tags loaded successfully", "count", len(tags))
		}
	}

	apis := handle.NewTagsHandler(s.logger, s.pap, s.auth.Authorizer())

	// the tag routes identify the caller once, up front; the service decides per operation.
	identify := authfiber.Identify(s.auth.Authorizer(), s.logger)

	// tags CRUD.
	group.Get(handle.PathTags, identify, apis.GetTags).
		Get(handle.PathTag, identify, apis.GetTag).
		Put(handle.PathTag, identify, apis.PutTag).
		Post(handle.PathTag, identify, apis.PostTag).
		Delete(handle.PathTag, identify, apis.DeleteTag)
}

func (s *Services) initPolicies(group fiber.Router) {
	apis := handle.NewPoliciesHandler(s.logger, s.pap, s.auth.Authorizer())

	// the policy routes identify the caller once, up front; the service decides per operation.
	identify := authfiber.Identify(s.auth.Authorizer(), s.logger)

	// policies CRUD.
	group.Get(handle.PathPolicies, identify, apis.GetPolicies).
		Get(handle.PathPolicy, identify, apis.GetPolicy).
		Put(handle.PathPolicy, identify, apis.PutPolicy).
		Post(handle.PathPolicy, identify, apis.PostPolicy).
		Delete(handle.PathPolicy, identify, apis.DeletePolicy)
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

	// the deployment routes identify the caller once, up front; the service decides per operation.
	identify := authfiber.Identify(s.auth.Authorizer(), s.logger)

	// deployments CRUD.
	group.Get(handle.PathStatuses, identify, apis.GetStatuses).
		Get(handle.PathCompressionTypes, identify, apis.GetCompressTypes).
		Get(handle.PathConfigs, identify, apis.GetConfigs).
		Get(handle.PathDeployments, identify, apis.GetDeployments).
		Get(handle.PathLastDeployment, identify, apis.GetLastDeployment).
		Get(handle.PathDeploymentID, identify, apis.GetDeployment).
		Post(handle.PathDeployment, identify, apis.PostDeployment)
}

// initBundleRoutes sets up the routing table for bundle retrieval requests.
func (s *Services) initBundleRoutes(_ context.Context, svc *fiber.App) {
	apis := handle.NewBundlesHandler(s.logger, s.pap, s.bundleManager, s.auth.Authorizer())

	// bundle retrieval for PDPs: an app reading the last deployment.
	v1 := svc.Group(handle.PathV1)
	v1.Get(handle.PathBundle, authfiber.Identify(s.auth.Authorizer(), s.logger), apis.GetBundle)
}

// initHealthRoutes sets up the routing table for health requests.
func (s *Services) initHealthRoutes(_ context.Context, svc *fiber.App) {
	// liveness and readiness.
	svc.Get(handle.PathHealthZ, s.chk.HealthZ).
		Get(handle.PathLiveZ, s.chk.HealthZ).
		Get(handle.PathReadyZ, s.chk.HealthZ)
}
