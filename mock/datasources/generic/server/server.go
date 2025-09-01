package server

import (
	"context"
	"log/slog"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/reader"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store/memory"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/generic/config"
)

func NewService(cfg *config.Config, logger *slog.Logger) (server.Service, error) {
	db := memory.New(nil)

	if err := reader.LoadFromPath(db, cfg.DataPath); err != nil {
		logger.Error("failed to load dataspace definition(s)", "path", cfg.DataPath, "error", err)
		return nil, err
	}

	s := &service{ctx: context.Background(), cfg: cfg, logger: logger, db: db, chk: handle.NewChecks()}

	// we're good to go once the service is running.
	s.chk.SetHealth(true)
	s.chk.SetAlive(true)
	s.chk.SetReady(true)

	opts := []server.Option{
		server.WithDefaults(),
		server.WithHostPort(cfg.Host, cfg.Port),
		server.WithAppName(config.AppName),
		server.WithTLS(cfg.CA, cfg.Cert, cfg.Key),
		server.WithTimeouts(cfg.ReadTimeout, cfg.WriteTimeout, cfg.IdleTimeout),
		server.WithMaxBody(cfg.MaxBody),
		server.WithRecovery(),
		server.WithSecurity(),
		server.WithCORS(cfg.CorsOrigins, cfg.CorsHeaders),
	}

	if cfg.Cert != "" {
		opts = append(opts, server.WithTLS(cfg.CA, cfg.Cert, cfg.Key))
	}

	s.Service = fiber.New(logger, s.initRoutes, opts...)

	return s, nil
}

type service struct {
	server.Service
	ctx    context.Context
	cfg    *config.Config
	logger *slog.Logger
	db     store.Storage
	chk    *handle.Checks
}
