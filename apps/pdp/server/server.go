// Package server handles the HTTP service component of the PDP.
package server

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server/fiber"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pdp/config"
)

// NewService initializes the HTTP service (implemented with fiber & fasthttp).
func NewService(cfg *config.Config, logger *slog.Logger) server.Service {
	s := &service{cfg: cfg, logger: logger}

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
}
