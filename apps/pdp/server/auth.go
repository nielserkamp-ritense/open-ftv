package server

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	handlers "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller/adl"
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
	decisionLog, err := s.newADL()
	if err != nil {
		s.logger.Error("failed to initialize authorization decision log", "error", err)
		return nil
	}

	var controller pdp.Controller
	if controller, err = s.newSelfAuthzController(decisionLog); err != nil || controller == nil {
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

func (s *Services) newSelfAuthzController(decisionLog *adl.ADL) (pdp.Controller, error) {
	controller, _, err := config.BuildSelfAuthzController(s.ctx, s.logger, s.l, &s.cfg.PIP, &s.cfg.Cerbos).
		WithBundledPAP(&s.cfg.PAP).
		WithOptions(pdp.WithADL(decisionLog)).
		Build()

	return controller, err
}

func (s *Services) newADL() (*adl.ADL, error) {
	svc := s.cfg.ADLService
	opts := []opentelemetry.Option{opentelemetry.WithBatchTimeout(time.Second * 5)}

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
