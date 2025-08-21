package server

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

// initRoutes sets up the routing table for HTTP requests.
func (s *service) initRoutes(ctx context.Context, svc *fiber.App) {
	s.ctx = ctx
	s.l = models.LanguageFromString(s.cfg.PAP.Language)

	s.auth = s.newAuth()
	if s.auth == nil {
		panic("failed to initialize authorization handler")
	}

	isPG := strings.EqualFold(s.cfg.Persist.Type, "postgres")

	var err error
	switch {
	case isPG:
		s.db, err = postgresql.New(s.ctx, s.cfg.Persist.PgURL, s.cfg.Persist.PgMaxLife, s.cfg.PgMaxConn)
	case s.cfg.Persist.Type != "":
		s.store, err = s.cfg.Persist.NewStore(ctx)
	}
	if err != nil {
		panic("failed to create persistence store: " + err.Error())
	}

	// initialize the PAP before the PIP, so database migrations happen before all else.
	if s.pap, err = s.newPAP(); err != nil {
		panic("failed to initialize PAP handler")
	}

	s.pip = s.newPIP()

	s.initHealth(svc)

	// API v1.
	v1 := svc.Group("/v1")
	s.initLanguages(v1)
	s.initTags(v1)
	s.initPolicies(v1)
	s.initAttributes(v1)
	s.initEntities(v1)

	if s.cfg.BundlePath != "" {
		s.initBundles(v1)
	}
}

func (s *service) initHealth(svc *fiber.App) {
	// liveness and readiness.
	svc.Get("/healthz", handle.HealthZ)
}

func (s *service) initLanguages(group fiber.Router) {
	apis := handle.NewLanguagesHandler(s.logger, s.pap, s.auth.Authorizer())

	// languages CRUD.
	group.Get(handle.PathLanguages, apis.GetLanguages)
}

func (s *service) initTags(group fiber.Router) {
	if tags := s.cfg.Tags(); len(tags) > 0 {
		if err := s.pap.LoadTags(tags); err != nil {
			s.logger.Error("failed to load tags", "error", err.Error())
		} else {
			s.logger.Info("tags loaded successfully", "count", len(tags))
		}
	}

	apis := handle.NewTagsHandler(s.logger, s.pap, s.auth.Authorizer())

	// tags CRUD.
	group.Get(handle.PathTags, apis.GetTags)
}

func (s *service) initPolicies(group fiber.Router) {
	apis := handle.NewPoliciesHandler(s.logger, s.pap, s.auth.Authorizer())

	// policies CRUD.
	group.Get(handle.PathPolicies, apis.GetPolicies)
	group.Get(handle.PathPolicy, apis.GetPolicy)
	group.Put(handle.PathPolicy, apis.PutPolicy)
	group.Post(handle.PathPolicy, apis.PostPolicy)
	group.Delete(handle.PathPolicy, apis.DeletePolicy)
}

func (s *service) initAttributes(group fiber.Router) {
	apis := handle.NewAttributesHandler(s.logger, s.pip, s.auth.Authorizer())

	// attributes CRUD.
	group.Get(handle.PathAttributes, apis.GetAttributes)
	group.Get(handle.PathAttribute, apis.GetAttribute)
	group.Put(handle.PathAttribute, apis.PutAttribute)
	group.Post(handle.PathAttribute, apis.PostAttribute)
	group.Delete(handle.PathAttribute, apis.DeleteAttribute)
}

func (s *service) initEntities(group fiber.Router) {
	apis := handle.NewEntitiesHandler(s.logger, s.pip, s.auth.Authorizer())

	// entities CRUD.
	group.Get(handle.PathEntities, apis.GetEntities)
	group.Get(handle.PathEntity, apis.GetEntity)
	group.Put(handle.PathEntity, apis.PutEntity)
	group.Post(handle.PathEntity, apis.PostEntity)
	group.Delete(handle.PathEntity, apis.DeleteEntity)
}

func (s *service) initBundles(group fiber.Router) {
	manager := bundles.NewManager(
		s.ctx,
		s.logger,
		bundles.WithConfig(s.cfg.BundlePath, s.cfg.BundleRecurse),
		bundles.WithPolicyLister(s.pap),
		bundles.WithAttributeLister(s.pip),
		bundles.WithEntityLister(s.pip),
		// bundles.WithRelationLister(s.pip),
		bundles.MaxWorkers(s.cfg.Workers),
		bundles.WithStageDelay(s.cfg.StageDelay),
		bundles.BundleTimeout(s.cfg.BundleTimeout),
	)

	apis := handle.NewBundlesHandler(s.logger, s.pap, manager, s.auth.Authorizer())

	// restart the last interrupted bundle deployment run if needed.
	time.AfterFunc(5*time.Second, func() { s.pap.RestartDeployment(manager) })

	// bundles CRUD.
	group.Get(handle.PathStatuses, apis.GetStatuses)
	group.Get(handle.PathCompressionTypes, apis.GetCompressTypes)
	group.Get(handle.PathConfigs, apis.GetConfigs)
	group.Get(handle.PathDeployments, apis.GetDeployments)
	group.Get(handle.PathLastDeployment, apis.GetLastDeployment)
	group.Get(handle.PathDeploymentID, apis.GetDeployment)
	group.Post(handle.PathDeployment, apis.PostDeployment)
}
