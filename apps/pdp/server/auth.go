package server

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	handlers "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions"
	cerbos_api "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cerbos-api"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller/adl"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/selfauthz"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/opentelemetry"
)

// authHandler implements the interface for handling authorization requests.
type authHandler struct {
	logger     *slog.Logger                    // application logger.
	controller pdp.Controller                  // generic PDP controller.
	zen        *handlers.AuthZENAuthorizer     // AuthZEN API.
	authorizer authorization.RequestAuthorizer // UI & bundle authorization.
}

func (s *Services) newAuth(basePath string) (*authHandler, error) {
	decisionLog, err := s.newADL()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize authorization decision log: %w", err)
	}

	var controller pdp.Controller
	if controller, err = s.newSelfAuthzController(decisionLog); err != nil {
		return nil, fmt.Errorf("failed to initialize pdp controller: %w", err)
	}

	if controller == nil {
		return nil, errors.New("failed to initialize pdp controller: no controller")
	}

	// var authenticator authentication.Authenticator
	// if authenticator, err = authentication.New(s.ctx, s.logger, s.cfg.Authentication.Type, func(uid string) (*models.Entity, uint64, error) {
	// 	return controller.GetPIP().GetEntity(uid)
	// }); err != nil {
	// 	return nil, fmt.Errorf("failed to initialize authenticator: %w", err)
	// }

	var authorizer authorization.RequestAuthorizer
	if authorizer, err = authorization.NewForController(controller, nil, s.cfg.Authenticate, s.cfg.FailClosedOnEmpty); err != nil {
		return nil, fmt.Errorf("failed to initialize authorizer: %w", err)
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
	}, nil
}

func (s *Services) newSelfAuthzController(decisionLog *adl.ADL) (pdp.Controller, error) {
	ip, err := pip.NewSelfAuthz(s.ctx, s.logger, s.cfg.PIP.Store, s.cfg.PIP.StoreRecurse, s.cfg.PullConfigs)
	if err != nil {
		return nil, err
	}

	return selfauthz.BuildController(s.ctx, s.logger, s.l, ip, (*cerbos_api.Config)(&s.cfg.Cerbos)).
		WithBundledPAP(s.cfg.PAP.Language, s.cfg.PAP.Store, s.cfg.PAP.StoreRecurse).
		WithOptions(pdp.WithADL(decisionLog)).
		Build()
}

func (s *Services) newADL() (*adl.ADL, error) {
	svc := s.cfg.ADLService

	const otelTimeout = time.Second * 5

	opts := []opentelemetry.Option{opentelemetry.WithBatchTimeout(otelTimeout)}

	// The ADL schema is owned by eam/log/decisions, not by an app: every app that uses the ADL
	// migrates it. See docs/adr/0004-adl-schema-migrated-by-every-app-that-uses-it.md.
	if err := decisions.Migrate(s.cfg.ADLDatabaseURL, s.logger); err != nil {
		return nil, fmt.Errorf("failed to migrate the authorization decision log: %w", err)
	}

	logger, err := decisions.NewWithPostgreSQL(s.ctx, svc, s.cfg.ADLDatabaseURL, nil, DBMaxConnLifetime, DBMaxConnections, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect the authorization decision log: %w", err)
	}

	instanceID := adlInstanceID()

	s.logger.Info("authorization decision log initialized", "type", "postgres", "service", svc, "instance", instanceID)
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
