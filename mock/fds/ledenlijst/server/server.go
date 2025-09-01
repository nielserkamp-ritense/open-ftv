package server

import (
	"context"
	"log/slog"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/fds/ledenlijst/config"
)

func NewService(cfg *config.Config, logger *slog.Logger) server.Service {
	s := &service{ctx: context.Background(), cfg: cfg, logger: logger, chk: handle.NewChecks()}

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
	)

	return s
}

type service struct {
	server.Service
	ctx    context.Context
	cfg    *config.Config
	logger *slog.Logger
	chk    *handle.Checks
}
