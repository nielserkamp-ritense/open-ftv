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
	s.l = models.LanguageFromString(s.cfg.Language)

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
		s.store, err = s.cfg.NewStore(ctx)
	}

	if err != nil {
		panic("failed to create persistence store: " + err.Error())
	}

	// initialize the PAP before the PIP, so database migrations happen before all else.
	if s.pap, err = s.newPAP(); err != nil {
		panic("failed to initialize PAP")
	}

	// seed the bundled cedar authorization policies into the (postgres) store with UUID ids
	// on an empty store, so they are enforced AND visible/editable in the UI. MUST run before
	// newAuth: the embedded PDP loads them at construction, before the authorizer decides
	// whether to run in NoAuth (fail-open) mode for an empty policy set.
	s.seedAuthzPolicies()

	// initialize authorization after the PAP, so the embedded self-authorization PDP
	// shares the very same (postgres-backed) policy store as the UI-managed policies.
	s.auth = s.newAuth()
	if s.auth == nil {
		panic("failed to initialize authorization manager")
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
		Patch(handle.PathPolicyStatus, apis.PatchPolicyStatus).
		Delete(handle.PathPolicy, apis.DeletePolicy)

	// policy versions.
	group.Get(handle.PathPolicyVersions, apis.GetPolicyVersions).
		Get(handle.PathPolicyVersion, apis.GetPolicyVersion).
		Post(handle.PathPolicyRestore, apis.PostPolicyRestore)
}

func (s *Services) initAttributes(group fiber.Router) {
	apis := handle.NewAttributesHandler(s.logger, s.pip, s.auth.Authorizer())

	// attributes CRUD.
	group.Get(handle.PathAttributes, apis.GetAttributes).
		Get(handle.PathAttribute, apis.GetAttribute).
		Put(handle.PathAttribute, apis.PutAttribute).
		Post(handle.PathAttribute, apis.PostAttribute).
		Patch(handle.PathAttributeStatus, apis.PatchAttributeStatus).
		Delete(handle.PathAttribute, apis.DeleteAttribute)

	// attribute versions.
	group.Get(handle.PathAttributeVersions, apis.GetAttributeVersions).
		Get(handle.PathAttributeVersion, apis.GetAttributeVersion).
		Post(handle.PathAttributeRestore, apis.PostAttributeRestore)
}

func (s *Services) initEntities(group fiber.Router) {
	apis := handle.NewEntitiesHandler(s.logger, s.pip, s.auth.Authorizer())

	// entities CRUD.
	group.Get(handle.PathEntities, apis.GetEntities).
		Get(handle.PathEntity, apis.GetEntity).
		Put(handle.PathEntity, apis.PutEntity).
		Post(handle.PathEntity, apis.PostEntity).
		Patch(handle.PathEntityStatus, apis.PatchEntityStatus).
		Delete(handle.PathEntity, apis.DeleteEntity)

	// entity versions.
	group.Get(handle.PathEntityVersions, apis.GetEntityVersions).
		Get(handle.PathEntityVersion, apis.GetEntityVersion).
		Post(handle.PathEntityRestore, apis.PostEntityRestore)
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
	opts := []bundles.Option{
		bundles.WithConfig(s.cfg.BundlePath, s.cfg.BundleRecurse),
		bundles.WithPolicyHandler(s.pap),
		bundles.WithDataHandler(s.pip),
		bundles.MaxWorkers(s.cfg.Workers),
		bundles.WithStageDelay(s.cfg.StageDelay),
		bundles.BundleTimeout(s.cfg.BundleTimeout),
	}

	last, _ := s.pap.LastDeployment()
	if last == nil {
		opts = append(opts, bundles.BootstrapDeployment())
	}

	s.bundleManager = bundles.NewManager(s.ctx, s.logger, opts...)

	apis := handle.NewBundlesHandler(s.logger, s.pap, s.bundleManager, s.auth.Authorizer())

	if last == nil {
		// create the first deployment, so any PDP can find this bundle at startup.
		time.AfterFunc(15*time.Second, func() {
			s.logger.Info("Initializing initial bootstrap deployment")

			if _, err := s.pap.NewDeployment(
				"Bootstrap",
				"Initial bootstrapped deployment with dummy policies",
				s.bundleManager,
				"*SYSTEM*",
			); err != nil {
				s.logger.Error("failed to initialize bootstrap deployment", "error", err.Error())
			}
		})
	} else {
		// restart the last interrupted bundle deployment run if needed.
		time.AfterFunc(15*time.Second, func() {
			s.pap.RestartDeployment(s.bundleManager)
		})
	}

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

	var (
		searcher search.Searcher
		err      error
	)

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
