// Package server handles the HTTP service component of the FSC Auth plugin.
package server

import (
	"context"
	"log/slog"

	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config"
	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

// NewService initializes the HTTP service (implemented with fiber & fasthttp).
func NewService(cfg *config.Config, logger *slog.Logger) server.Service {
	s := &service{cfg: cfg, logger: logger, chk: handle.NewChecks()}

	// we're good to go once the service is running.
	s.chk.SetHealth(true)
	s.chk.SetAlive(true)
	s.chk.SetReady(true)

	s.Service = fiber.New(
		logger,
		s.initRoutes,
		server.WithDefaults(),
		server.WithHostPort(cfg.Host, cfg.Port),
		server.WithAppName(config.AppName),
		server.WithTLS(cfg.CA, cfg.Cert, cfg.Key),
		server.WithTimeouts(cfg.ReadTimeout, cfg.WriteTimeout, cfg.IdleTimeout),
		server.WithMaxBody(cfg.MaxBody),
		server.WithRecovery(),
		server.WithSecurity(),
		server.WithCORS(cfg.CorsOrigins, cfg.CorsHeaders),
	)

	return s
}

type service struct {
	server.Service
	ctx    context.Context
	cfg    *config.Config
	logger *slog.Logger
	auth   AuthHandler
	l      models.Language
	store  store.Store
	db     *postgresql.Postgres
	pap    *pap.PAP
	pip    *pip.PIP
	chk    *handle.Checks
}
