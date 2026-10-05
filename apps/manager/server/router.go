package server

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	authfiber "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions"
	searchPG "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/search/postgresql"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/principals"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/settings"
)

// initRoutes sets up the routing table for HTTP requests.
func (s *Services) initRoutes(ctx context.Context, svc *fiber.App) {
	s.ctx = ctx

	s.l = models.LanguageFromString(s.cfg.Language)

	// initialize the PAP before the PIP, so database migrations happen before all else.
	var err error
	if s.pap, err = s.newPAP(); err != nil {
		s.logger.Error("failed to initialize PAP", "error", err)
		panic("failed to initialize PAP: " + err.Error())
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
		s.logger.Error("failed to initialize authorization manager")
		panic("failed to initialize authorization manager")
	}

	if s.pip, err = s.newPIP(); err != nil {
		s.logger.Error("failed to initialize PIP", "error", err)
		panic("failed to initialize PIP: " + err.Error())
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
// resolved to names.
func (s *Services) principalOption() []handle.HandlerOption {
	return []handle.HandlerOption{handle.WithPrincipals(principals.NewDBWithPool(s.db))}
}

func (s *Services) initLanguages(group fiber.Router) {
	apis := handle.NewLanguagesHandler(s.logger, s.pap, s.auth.Authorizer(), s.principalOption()...)

	// the language routes identify the caller once, up front; the service decides per operation.
	identify := authfiber.Identify(s.auth.Authorizer(), s.logger)

	// languages CRUD.
	group.Get(handle.PathLanguages, identify, apis.GetLanguages)
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

	apis := handle.NewTagsHandler(s.logger, s.pap, s.auth.Authorizer(), s.principalOption()...)

	// the tag routes identify the caller once, up front; the service decides per operation.
	identify := authfiber.Identify(s.auth.Authorizer(), s.logger)

	// tags CRUD.
	group.Get(handle.PathTags, identify, apis.GetTags).
		Get(handle.PathTag, identify, apis.GetTag).
		Put(handle.PathTag, identify, apis.PutTag).
		Post(handle.PathTag, identify, apis.PostTag).
		Delete(handle.PathTag, identify, apis.DeleteTag)
}

func (s *Services) initSettings(group fiber.Router) {
	settingsDB := settings.NewSettingsDBWithPool(s.db)

	apis := handle.NewSettingsHandler(s.logger, settingsDB, s.auth.Authorizer(), s.principalOption()...)

	// the settings routes identify the caller once, up front; the service decides per operation.
	identify := authfiber.Identify(s.auth.Authorizer(), s.logger)

	group.Get(handle.PathSettings, identify, apis.GetSettings).
		Put(handle.PathSettings, identify, apis.PutSettings)
}

func (s *Services) initPolicies(group fiber.Router) {
	apis := handle.NewPoliciesHandler(s.logger, s.pap, s.auth.Authorizer(), s.principalOption()...)

	// the policy routes identify the caller once, up front; the service decides per operation.
	identify := authfiber.Identify(s.auth.Authorizer(), s.logger)

	// policies CRUD.
	group.Get(handle.PathPolicies, identify, apis.GetPolicies).
		Get(handle.PathPolicy, identify, apis.GetPolicy).
		Put(handle.PathPolicy, identify, apis.PutPolicy).
		Post(handle.PathPolicy, identify, apis.PostPolicy).
		Patch(handle.PathPolicyStatus, identify, apis.PatchPolicyStatus).
		Delete(handle.PathPolicy, identify, apis.DeletePolicy)

	// policy versions.
	group.Get(handle.PathPolicyVersions, identify, apis.GetPolicyVersions).
		Get(handle.PathPolicyVersion, identify, apis.GetPolicyVersion).
		Post(handle.PathPolicyRestore, identify, apis.PostPolicyRestore)
}

func (s *Services) initAttributes(group fiber.Router) {
	apis := handle.NewAttributesHandler(s.logger, s.pip, s.auth.Authorizer(), s.principalOption()...)

	// the attribute routes identify the caller once, up front; the service decides per operation.
	identify := authfiber.Identify(s.auth.Authorizer(), s.logger)

	// attributes CRUD.
	group.Get(handle.PathAttributes, identify, apis.GetAttributes).
		Get(handle.PathAttribute, identify, apis.GetAttribute).
		Put(handle.PathAttribute, identify, apis.PutAttribute).
		Post(handle.PathAttribute, identify, apis.PostAttribute).
		Patch(handle.PathAttributeStatus, identify, apis.PatchAttributeStatus).
		Delete(handle.PathAttribute, identify, apis.DeleteAttribute)

	// attribute versions.
	group.Get(handle.PathAttributeVersions, identify, apis.GetAttributeVersions).
		Get(handle.PathAttributeVersion, identify, apis.GetAttributeVersion).
		Post(handle.PathAttributeRestore, identify, apis.PostAttributeRestore)
}

func (s *Services) initEntities(group fiber.Router) {
	apis := handle.NewEntitiesHandler(s.logger, s.pip, s.auth.Authorizer(), s.principalOption()...)

	// the entity routes identify the caller once, up front; the service decides per operation.
	identify := authfiber.Identify(s.auth.Authorizer(), s.logger)

	// entities CRUD.
	group.Get(handle.PathEntities, identify, apis.GetEntities).
		Get(handle.PathEntity, identify, apis.GetEntity).
		Put(handle.PathEntity, identify, apis.PutEntity).
		Post(handle.PathEntity, identify, apis.PostEntity).
		Patch(handle.PathEntityStatus, identify, apis.PatchEntityStatus).
		Delete(handle.PathEntity, identify, apis.DeleteEntity)

	// entity versions.
	group.Get(handle.PathEntityVersions, identify, apis.GetEntityVersions).
		Get(handle.PathEntityVersion, identify, apis.GetEntityVersion).
		Post(handle.PathEntityRestore, identify, apis.PostEntityRestore)
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

func (s *Services) initADL(group fiber.Router) {
	// The ADL schema is owned by eam/log/decisions, not by an app: a manager without a
	// PDP alongside it (in-app deployment) would otherwise never see the schema created,
	// and with a PDP alongside it startup order would decide whether the log works.
	// See docs/adr/0004-adl-schema-migrated-by-every-app-that-uses-it.md.
	err := decisions.Migrate(s.cfg.ADLDatabaseURL, s.logger)
	if err != nil {
		s.logger.Error("failed to migrate the Authorization Decision Log database", "error", err)
		panic("failed to migrate the Authorization Decision Log database: " + err.Error())
	}

	searcher, err := searchPG.New(s.ctx, s.cfg.ADLDatabaseURL, nil, DBMaxConnLifetime, DBMaxConnections)
	if err != nil {
		s.logger.Error("failed to initialize Authorization Decision Log search API", "error", err)
		panic("failed to initialize Authorization Decision Log search API: " + err.Error())
	}

	adl := handle.NewADLHandler(s.logger, searcher, s.auth.Authorizer())
	s.logger.Info("adl initialized")

	grp2 := group.Group(handle.PathADL)
	// the ADL routes identify the caller once, up front; the service decides per operation.
	grp2.Get(handle.PathEntries, authfiber.Identify(s.auth.Authorizer(), s.logger), adl.Search)
}

// initBundleRoutes sets up the routing table for bundle retrieval requests.
func (s *Services) initBundleRoutes(_ context.Context, svc *fiber.App) {
	apis := handle.NewBundlesHandler(s.logger, s.pap, s.bundleManager, s.auth.Authorizer(), s.principalOption()...)

	// bundle retrieval for PDPs: an app reading the last deployment.
	v1 := svc.Group(handle.PathV1)
	v1.Get(handle.PathBundleID, authfiber.Identify(s.auth.Authorizer(), s.logger), apis.GetBundle)
}

// initHealthRoutes sets up the routing table for health requests.
func (s *Services) initHealthRoutes(_ context.Context, svc *fiber.App) {
	// liveness and readiness.
	svc.Get(handle.PathHealthZ, s.chk.HealthZ).
		Get(handle.PathLiveZ, s.chk.HealthZ).
		Get(handle.PathReadyZ, s.chk.HealthZ)
}
