// Package server handles the HTTP service component of the FSC Auth plugin.
package server

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pip/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/warc"
)

// NewService initializes the HTTP service (implemented with fiber & fasthttp).
func NewService(cfg *config.Config, logger *slog.Logger) server.Service {
	s := &service{cfg: cfg, logger: logger}

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
	l      models.Language
	auth   AuthHandler
	pip    pip.PIP
	warc   *warc.Writer
}
