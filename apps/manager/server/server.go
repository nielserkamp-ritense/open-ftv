// Package server handles the HTTP service component of the FSC Auth plugin.
package server

import (
	"context"
	"log/slog"
	"sync"

	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server"
	eam_fiber "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

// NewExternal initializes the HTTP(S) services (implemented with fiber & fasthttp).
func NewExternal(cfg *config.Config, logger *slog.Logger) *Services {
	s := &Services{cfg: cfg, logger: logger, chk: handle.NewChecks()}

	// we're good to go once the service is running.
	s.chk.SetHealth(true)
	s.chk.SetAlive(true)
	s.chk.SetReady(true)

	// main service
	opts := []server.Option{
		server.WithDefaults(),
		server.WithHostPort(cfg.Host, cfg.Port),
		server.WithAppName(config.AppName),
		server.WithSvcName("main"),
		server.WithTimeouts(cfg.ReadTimeout, cfg.WriteTimeout, cfg.IdleTimeout),
		server.WithMaxBody(cfg.MaxBody),
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

	s.main = eam_fiber.New(logger, s.initRoutes, opts...)

	if s.cfg.BundlePath != "" {
		// internal service (bundles)
		opts = []server.Option{
			server.WithDefaults(),
			server.WithHostPort(cfg.InternalHost, cfg.InternalPort),
			server.WithAppName(config.AppName),
			server.WithSvcName("internal"),
			server.WithTimeouts(cfg.InternalRead, cfg.InternalWrite, cfg.InternalIdle),
			server.WithMaxBody(cfg.InternalMaxBody),
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

		s.bundles = eam_fiber.New(logger, s.initBundleRoutes, opts...)
	}

	// health service
	s.health = eam_fiber.New(
		logger,
		s.initHealthRoutes,
		server.WithDefaults(),
		server.WithHostPort(cfg.HealthHost, cfg.HealthPort),
		server.WithAppName(config.AppName),
		server.WithSvcName("health"),
		server.WithTimeouts(cfg.HealthRead, cfg.HealthWrite, cfg.HealthIdle),
		server.WithMaxBody(cfg.HealthMaxBody),
		server.WithRecovery(),
		server.WithSecurity(),
		server.WithCORS(cfg.HealthOrigins, cfg.HealthHeaders),
	)

	return s
}

// Serve activates the HTTP(S) services.
func (s *Services) Serve() {
	wg := sync.WaitGroup{}
	wg.Add(2)

	if s.cfg.BundlePath != "" {
		wg.Add(1)
		go s.bundles.ServeWithWG(&wg)
	}

	go s.main.ServeWithWG(&wg)
	go s.health.ServeWithWG(&wg)

	wg.Wait()
}

// Shutdown shuts down the HTTP(S) services.
func (s *Services) Shutdown() {
	s.health.Shutdown()
	s.main.Shutdown()

	if s.cfg.BundlePath != "" {
		s.bundles.Shutdown()
	}
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
	store         store.Store
	db            *postgresql.Postgres
	pap           *pap.PAP
	pip           *pip.PIP
	chk           *handle.Checks
}

func (s *Services) GetMainService() server.Service {
	return s.main
}

func (s *Services) GetHealthService() server.Service {
	return s.health
}
