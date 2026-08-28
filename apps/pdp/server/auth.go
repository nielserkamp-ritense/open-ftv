package server

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	handlers "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	cedar_embedded "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	cerbos_api "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cerbos-api"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller/adl"
	opa_embedded "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/opa-embedded"
	openfga_embedded "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/openfga-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/opentelemetry"
)

// authHandler implements the interface for handling authorization requests.
type authHandler struct {
	logger     *slog.Logger                // application logger.
	controller pdp.Controller              // generic PDP controller.
	zen        *handlers.AuthZENAuthorizer // AuthZEN API.
	authorizer authorization.Authorizer    // UI & bundle authorization.
}

func (s *Services) newAuth(basePath string) *authHandler {
	var err error

	var decisionLog *adl.ADL
	if lt := s.cfg.DecisionLog.Type; lt != "" {
		if decisionLog, err = s.newADL(lt); err != nil || decisionLog == nil {
			s.logger.Error("failed to initialize authorization decision log", "error", err)
			return nil
		}
	}

	var controller pdp.Controller
	if controller, err = s.newController(decisionLog); err != nil || controller == nil {
		s.logger.Error("failed to initialize pdp controller", "error", err)
		return nil
	}

	// var authenticator authentication.Authenticator
	// if authenticator, err = s.cfg.Authentication.NewAuthenticator(s.ctx, s.logger, func(uid string) (*models.Entity, uint64, error) {
	// 	return controller.GetPIP().GetEntity(uid)
	// }); err != nil {
	// 	s.logger.Error("failed to initialize authenticator", "error", err)
	// 	return nil
	// }

	var authorizer authorization.Authorizer
	if authorizer, err = s.cfg.NewAuthorizer(controller, nil); err != nil {
		s.logger.Error("failed to initialize authorizer", "error", err)
		return nil
	}

	var prefix string
	if s.cfg.AuthZENMethod != "" || s.cfg.AuthZENDomain != "" {
		prefix = fmt.Sprintf("%s://%s%s", s.cfg.AuthZENMethod, s.cfg.AuthZENDomain, basePath)
	}

	zen := handlers.NewAuthHandlerZEN(s.logger, decisionLog, controller, prefix).
		WithEvaluations().
		WithSearchSubject().
		WithSearchAction().
		WithSearchResource()

	return &authHandler{
		logger:     s.logger,
		controller: controller,
		zen:        zen,
		authorizer: authorizer,
	}
}

func (s *Services) newController(decisionLog *adl.ADL) (pdp.Controller, error) {
	ep := pep.New(s.ctx, s.logger)

	ip, err := s.cfg.NewSelfAuthzPIP(s.ctx, s.logger, s.l)
	if err != nil {
		return nil, err
	}

	ap, err2 := s.cfg.NewSelfAuthzPAP(s.ctx, s.logger)
	if err2 != nil {
		return nil, err2
	}

	options := []pdp.Option{
		pdp.WithContext(s.ctx),
		pdp.WithLogger(s.logger),
		pdp.WithPEP(ep),
		pdp.WithPIP(ip),
		pdp.WithPAP(ap),
		pdp.WithADL(decisionLog),
	}

	switch s.l {
	case models.CEDAR:
		return cedar_embedded.NewController(options...), nil
	case models.REGO:
		return opa_embedded.NewController(options...), nil
	case models.OPENFGA:
		return openfga_embedded.NewController(options...), nil
	case models.CERBOS:
		cerbosCFG := cerbos_api.Config{Addr1: s.cfg.Address, Addr2: s.cfg.AdminAddress, CA: s.cfg.CA, User: s.cfg.User, Pswd: s.cfg.Pswd}
		return cerbos_api.NewController(cerbosCFG, options...), nil
	default:
		return nil, fmt.Errorf("unsupported policy language '%s'", s.cfg.Language)
	}
}

func (s *Services) newADL(lt string) (*adl.ADL, error) {
	cfg := s.cfg.DecisionLog
	svc := cfg.Service
	opts := []opentelemetry.Option{opentelemetry.WithBatchTimeout(cfg.Timeout)}

	var pg bool

	switch strings.ToLower(lt) {
	case "pg", "postgres", "postgresql":
		pg = true
	case "ot", "otel", "opentelemetry":
		opts = append(opts, opentelemetry.WithOT(cfg.OtelURL, cfg.OtelInsecure))
	case "slog":
		opts = append(opts, opentelemetry.WithSLog(s.logger, cfg.SlogMsg))
	case "stdout":
		opts = append(opts, opentelemetry.WithFile(os.Stdout, cfg.Pretty))
	case "stderr":
		opts = append(opts, opentelemetry.WithFile(os.Stderr, cfg.Pretty))
	default:
		return nil, fmt.Errorf("unsupported decision log type '%s'", cfg.Type)
	}

	var (
		logger *decisions.Logger
		err    error
	)

	if pg {
		if err = s.checkMigrations(); err == nil {
			logger, err = decisions.NewWithPostgreSQL(s.ctx, svc, cfg.PgURL, nil, cfg.PgMaxLife, cfg.PgMaxConn, opts...)
		}
	} else {
		logger, err = decisions.New(s.ctx, svc, opts...)
	}

	if err != nil {
		return nil, err
	}

	instanceID := adlInstanceID()

	s.logger.Info("authorization decision log initialized", "type", lt, "service", svc, "instance", instanceID)
	s.decisionLog = adl.New(logger, adl.WithResource(map[string]any{
		"organization": svc,
		"instanceId":   instanceID,
	}))

	return s.decisionLog, nil
}

// adlInstanceID returns a "pdp-" prefixed identifier for this process: the host/pod name if
// available, otherwise a random ID.
func adlInstanceID() string {
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		return "pdp-" + hostname
	}

	return "pdp-" + uuid.New().String()
}

// checkMigrations migrates the ADL database, which is the only database this app has.
//
// The ADL-specific settings (PDP_ADL_MIGRATE_*) take precedence; when they are not given the
// general migration settings (PDP_MIGRATE_*) are used, which is how this app has always been
// configured. See docs/adr/0004-adl-schema-migrated-by-every-app-that-uses-it.md.
func (s *Services) checkMigrations() error {
	adlCfg, cfg := s.cfg.DecisionLog, s.cfg.Migration

	// An explicit ADL source wins; on the default (embedded scripts) the general source is
	// used, so a PDP configured with PDP_MIGRATE_SOURCE=file://... keeps working unchanged.
	// An explicitly emptied ADL source switches migration off.
	source := adlCfg.MigrateSource
	if strings.EqualFold(source, "*embed*") && cfg.Source != "" {
		source = cfg.Source
	}

	steps := adlCfg.MigrateSteps
	if steps == 0 {
		steps = cfg.Steps
	}

	return decisions.Migrate(source, adlCfg.PgURL, steps, adlCfg.MigrateAuto || cfg.Auto, s.logger)
}
