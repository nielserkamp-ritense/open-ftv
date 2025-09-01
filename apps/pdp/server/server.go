// Package server handles the HTTP service component of the PDP.
package server

import (
	"context"
	"log/slog"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config"
)

// NewService initializes the HTTP service (implemented with fiber & fasthttp).
func NewService(cfg *config.Config, logger *slog.Logger) server.Service {
	s := &service{cfg: cfg, logger: logger, chk: handle.NewChecks()}

	s.chk.SetHealth(true)
	s.chk.SetAlive(true)
	s.chk.SetReady(false) // wait for latest bundle first!

	s.Service = fiber.New(
		logger,
		s.initRoutes,
		server.WithDefaults(),
		server.WithHostPort(cfg.Server.Host, cfg.Server.Port),
		server.WithAppName(config.AppName),
		server.WithTLS(cfg.Server.CA, cfg.Server.Cert, cfg.Server.Key),
		server.WithTimeouts(cfg.Server.ReadTimeout, cfg.Server.WriteTimeout, cfg.Server.IdleTimeout),
		server.WithMaxBody(cfg.Server.MaxBody),
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
	l      models.Language
	auth   AuthHandler
	chk    *handle.Checks
}
