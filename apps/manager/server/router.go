package server

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/search"
	searchPG "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/search/postgresql"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

// initRoutes sets up the routing table for HTTP requests.
func (s *Services) initRoutes(ctx context.Context, svc *fiber.App) {
	s.ctx = ctx
	s.l = models.LanguageFromString(s.cfg.PAP.Language)

	s.auth = s.newAuth()
	if s.auth == nil {
		panic("failed to initialize authorization manager")
	}

	var isPG bool
	switch strings.ToLower(s.cfg.Persist.Type) {
	case "pg", "postgres", "postgresql":
		isPG = true
	}

	var err error
	switch {
	case isPG:
		s.db, err = postgresql.New(s.ctx, s.cfg.Persist.PgURL, s.cfg.Persist.PgMaxLife, s.cfg.Persist.PgMaxConn)
	case s.cfg.Persist.Type != "":
		s.store, err = s.cfg.Persist.NewStore(ctx)
	}
	if err != nil {
		panic("failed to create persistence store: " + err.Error())
	}

	// initialize the PAP before the PIP, so database migrations happen before all else.
	if s.pap, err = s.newPAP(); err != nil {
		panic("failed to initialize PAP")
	}

	s.pip = s.newPIP()
	if s.pip == nil {
		panic("failed to initialize PIP")
	}

	// API v1.
	v1 := svc.Group(handle.PathV1)
	s.initLanguages(v1)
	s.initTags(v1)
	s.initPolicies(v1)
	s.initAttributes(v1)
	s.initEntities(v1)
	s.initRelations(v1)
	s.initDeployments(v1)
	s.initADL(v1)
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

func (s *Services) initAttributes(group fiber.Router) {
	apis := handle.NewAttributesHandler(s.logger, s.pip, s.auth.Authorizer())

	// attributes CRUD.
	group.Get(handle.PathAttributes, apis.GetAttributes).
		Get(handle.PathAttribute, apis.GetAttribute).
		Put(handle.PathAttribute, apis.PutAttribute).
		Post(handle.PathAttribute, apis.PostAttribute).
		Delete(handle.PathAttribute, apis.DeleteAttribute)
}

func (s *Services) initEntities(group fiber.Router) {
	apis := handle.NewEntitiesHandler(s.logger, s.pip, s.auth.Authorizer())

	// entities CRUD.
	group.Get(handle.PathEntities, apis.GetEntities).
		Get(handle.PathEntity, apis.GetEntity).
		Put(handle.PathEntity, apis.PutEntity).
		Post(handle.PathEntity, apis.PostEntity).
		Delete(handle.PathEntity, apis.DeleteEntity)
}

func (s *Services) initRelations(_ fiber.Router) {
	// apis := handle.NewRelationsHandler(s.logger, s.pip, s.auth.Authorizer())
	//
	// // relations CRUD.
	// group.Get(handle.PathRelations, apis.GetRelations).
	//    Get(handle.PathRelation, apis.GetRelation).
	//    Put(handle.PathRelation, apis.PutRelation).
	//    Post(handle.PathRelation, apis.PostRelation).
	//    Delete(handle.PathRelation, apis.DeleteRelation)
}

func (s *Services) initDeployments(group fiber.Router) {
	s.bundleManager = bundles.NewManager(
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

func (s *Services) initADL(group fiber.Router) {
	cfg := s.cfg.DecisionLog

	var searcher search.Searcher
	var err error

	switch strings.ToLower(cfg.Type) {
	case "pg", "postgres", "postgresql":
		searcher, err = searchPG.New(s.ctx, cfg.PgURL, nil, cfg.PgMaxLife, cfg.PgMaxConn)
	default:
		return
	}

	if err != nil {
		s.logger.Error("failed to initialize Authorization Decision Log search API", "error", err.Error())
	}

	adl := handle.NewADLHandler(s.logger, searcher, s.auth.Authorizer())
	s.logger.Info("adl initialized", "type", cfg.Type)

	grp2 := group.Group(handle.PathADL)
	grp2.Get(handle.PathEntries, adl.Search)
}

// initBundleRoutes sets up the routing table for bundle retrieval requests.
func (s *Services) initBundleRoutes(_ context.Context, svc *fiber.App) {
	apis := handle.NewBundlesHandler(s.logger, s.pap, s.bundleManager, s.auth.Authorizer())

	// bundle retrieval for PDPs.
	v1 := svc.Group(handle.PathV1)
	v1.Get(handle.PathBundleID, apis.GetBundle)
}

// initHealthRoutes sets up the routing table for health requests.
func (s *Services) initHealthRoutes(_ context.Context, svc *fiber.App) {
	// liveness and readiness.
	svc.Get(handle.PathHealthZ, s.chk.HealthZ).
		Get(handle.PathLiveZ, s.chk.HealthZ).
		Get(handle.PathReadyZ, s.chk.HealthZ)
}
