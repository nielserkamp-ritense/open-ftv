// Package server handles the HTTP service component of the FSC Auth plugin.
package server

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pap/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

// Connection pool settings for the Postgres database this app uses.
const (
	DBMaxConnLifetime = 5 * time.Minute
	DBMaxConnections  = 100
)

// NewService initializes the HTTP(S) services (implemented with fiber & fasthttp).
func NewService(cfg *config.Config, logger *slog.Logger) *Services {
	s := &Services{
		cfg:    cfg,
		logger: logger,
		chk:    handle.NewChecks(),
	}

	// we're good to go once the service is running.
	s.chk.SetHealth(true)
	s.chk.SetAlive(true)
	s.chk.SetReady(true)

	dbCtx, dbCancel := context.WithCancel(context.Background())
	s.dbCancel = dbCancel

	postgresDB, err := postgresql.New(dbCtx, cfg.DatabaseURL, DBMaxConnLifetime, DBMaxConnections)
	if err != nil {
		dbCancel()
		logger.Error("failed to connect the postgres database", "error", err)
		panic("failed to connect the postgres database: " + err.Error())
	}

	s.db = postgresDB

	opts := []server.Option{
		server.WithDefaults(),
		server.WithHostPort(cfg.Host, cfg.Port),
		server.WithSvcName("main"),
		server.WithTimeouts(cfg.ReadTimeout, cfg.WriteTimeout, cfg.IdleTimeout),
		server.WithRecovery(),
		server.WithSecurity(),
		server.WithCORS(cfg.CorsOrigins, cfg.CorsHeaders),
	}

	if cfg.Cert != "" && cfg.Key != "" {
		opts = append(opts, server.WithTLS(cfg.CA, cfg.Cert, cfg.Key))
	}

	if cfg.CA != "" {
		opts = append(opts, server.WithMutualTLS())
	}

	s.main = fiber.New(logger, s.initMainRoutes, opts...)

	if s.cfg.BundlePath != "" {
		opts = []server.Option{
			server.WithDefaults(),
			server.WithHostPort(cfg.InternalHost, cfg.InternalPort),
			server.WithSvcName("internal"),
			server.WithTimeouts(cfg.InternalRead, cfg.InternalWrite, cfg.InternalIdle),
			server.WithRecovery(),
			server.WithSecurity(),
			server.WithCORS(cfg.InternalOrigins, cfg.InternalHeaders),
		}

		if cfg.Cert != "" && cfg.Key != "" {
			opts = append(opts, server.WithTLS(cfg.InternalCA, cfg.InternalCert, cfg.InternalKey))
		}

		if cfg.CA != "" {
			opts = append(opts, server.WithMutualTLS())
		}

		s.bundles = fiber.New(logger, s.initBundleRoutes, opts...)
	}

	s.health = fiber.New(
		logger,
		s.initHealthRoutes,
		server.WithDefaults(),
		server.WithHostPort(cfg.HealthHost, cfg.HealthPort),
		server.WithSvcName("health"),
		server.WithTimeouts(cfg.HealthRead, cfg.HealthWrite, cfg.HealthIdle),
		server.WithRecovery(),
		server.WithSecurity(),
		server.WithCORS(cfg.HealthOrigins, cfg.HealthHeaders),
	)

	return s
}

// Serve blocks until all services have stopped, and then closes the database connection pool.
func (s *Services) Serve() {
	wg := sync.WaitGroup{}

	if s.cfg.BundlePath != "" {
		s.serve(&wg, s.bundles)
	}

	s.serve(&wg, s.main)
	s.serve(&wg, s.health)

	wg.Wait()

	if s.dbCancel != nil {
		s.dbCancel() // closes the connection pool.
	}
}

// Shutdown stops the HTTP(S) services gracefully; Serve returns once they have stopped.
func (s *Services) Shutdown() {
	s.health.Shutdown()
	s.main.Shutdown()

	if s.cfg.BundlePath != "" {
		s.bundles.Shutdown()
	}
}

// serve runs svc until it stops. A service stops on an OS signal or when it fails to start; either way, the others stop too.
func (s *Services) serve(wg *sync.WaitGroup, svc server.Service) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		svc.Serve()
		s.Shutdown()
	}()
}

// Services contains the details of the HTTP(S) services.
type Services struct {
	ctx           context.Context
	logger        *slog.Logger
	cfg           *config.Config
	l             models.Language
	main          server.Service
	bundles       server.Service
	health        server.Service
	auth          AuthHandler
	bundleManager *bundles.Manager
	pap           *pap.PAP
	db            *postgresql.Postgres
	dbCancel      context.CancelFunc
	chk           *handle.Checks
}
