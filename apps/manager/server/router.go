package server

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/search"
	searchPG "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/search/postgresql"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/principals"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/settings"
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
	s.initSettings(v1)
	s.initPolicies(v1)
	s.initAttributes(v1)
	s.initEntities(v1)
	s.initRelations(v1)
	s.initDeployments(v1)
	s.initADL(v1)
}

// principalOption gives a handler the principal store, so attribution ids in its responses are
// resolved to names. It is empty when the manager runs without postgres, in which case responses
// carry ids only.
func (s *Services) principalOption() []handle.HandlerOption {
	if s.db == nil {
		return nil
	}

	return []handle.HandlerOption{handle.WithPrincipals(principals.NewDBWithPool(s.db))}
}

func (s *Services) initLanguages(group fiber.Router) {
	apis := handle.NewLanguagesHandler(s.logger, s.pap, s.auth.Authorizer(), s.principalOption()...)

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

	apis := handle.NewTagsHandler(s.logger, s.pap, s.auth.Authorizer(), s.principalOption()...)

	// tags CRUD.
	group.Get(handle.PathTags, apis.GetTags).
		Get(handle.PathTag, apis.GetTag).
		Put(handle.PathTag, apis.PutTag).
		Post(handle.PathTag, apis.PostTag).
		Delete(handle.PathTag, apis.DeleteTag)
}

func (s *Services) initSettings(group fiber.Router) {
	if s.db == nil {
		return
	}

	apis := handle.NewSettingsHandler(s.logger, settings.NewSettingsDBWithPool(s.db), s.auth.Authorizer(), s.principalOption()...)

	group.Get(handle.PathSettings, apis.GetSettings).
		Put(handle.PathSettings, apis.PutSettings)
}

func (s *Services) initPolicies(group fiber.Router) {
	apis := handle.NewPoliciesHandler(s.logger, s.pap, s.auth.Authorizer(), s.principalOption()...)

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
	apis := handle.NewAttributesHandler(s.logger, s.pip, s.auth.Authorizer(), s.principalOption()...)

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
	apis := handle.NewEntitiesHandler(s.logger, s.pip, s.auth.Authorizer(), s.principalOption()...)

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

	apis := handle.NewBundlesHandler(s.logger, s.pap, s.bundleManager, s.auth.Authorizer(), s.principalOption()...)

	if last == nil {
		// create the first deployment, so any PDP can find this bundle at startup.
		time.AfterFunc(15*time.Second, func() {
			s.logger.Info("Initializing initial bootstrap deployment")

			if _, err := s.pap.NewDeployment(
				"Bootstrap",
				"Initial bootstrapped deployment with dummy policies",
				s.bundleManager,
				identity.NewSystemPrincipal(),
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
		// The ADL schema is owned by eam/log/decisions, not by an app: a manager without a
		// PDP alongside it (in-app deployment) would otherwise never see the schema created,
		// and with a PDP alongside it startup order would decide whether the log works.
		// See docs/adr/0004-adl-schema-migrated-by-every-app-that-uses-it.md.
		if err = decisions.Migrate(cfg.MigrateSource, cfg.PgURL, s.adlMigrateSteps(), s.adlMigrateAuto(), s.logger); err != nil {
			panic("failed to migrate the Authorization Decision Log database: " + err.Error())
		}

		searcher, err = searchPG.New(s.ctx, cfg.PgURL, nil, cfg.PgMaxLife, cfg.PgMaxConn)
	default:
		return
	}

	if err != nil {
		panic("failed to initialize Authorization Decision Log search API: " + err.Error())
	}

	adl := handle.NewADLHandler(s.logger, searcher, s.auth.Authorizer())
	s.logger.Info("adl initialized", "type", cfg.Type)

	grp2 := group.Group(handle.PathADL)
	grp2.Get(handle.PathEntries, adl.Search)
}

// adlMigrateAuto reports whether the ADL must be migrated up to the latest level, falling back
// to the manager's general migration settings when the ADL-specific setting is not given.
func (s *Services) adlMigrateAuto() bool {
	// MigrateAuto is the ADL setting (ADL_MIGRATE_AUTO), Auto the general one (MIGRATE_AUTO).
	return s.cfg.MigrateAuto || s.cfg.Auto
}

// adlMigrateSteps returns the number of ADL migration steps, falling back to the manager's
// general migration settings when the ADL-specific setting is not given.
//
// NOTE: only the number of steps and the auto flag are inherited, never the source:
// MANAGER_MIGRATE_SOURCE locates the scripts of the manager's OWN database, which must never
// be applied to the ADL database. ADL_MIGRATE_SOURCE defaults to the embedded ADL scripts.
func (s *Services) adlMigrateSteps() int {
	// MigrateSteps is the ADL setting (ADL_MIGRATE_STEPS), Steps the general one (MIGRATE_STEPS).
	if steps := s.cfg.MigrateSteps; steps != 0 {
		return steps
	}

	return s.cfg.Steps
}

// initBundleRoutes sets up the routing table for bundle retrieval requests.
func (s *Services) initBundleRoutes(_ context.Context, svc *fiber.App) {
	apis := handle.NewBundlesHandler(s.logger, s.pap, s.bundleManager, s.auth.Authorizer(), s.principalOption()...)

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
