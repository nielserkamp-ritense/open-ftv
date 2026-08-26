// Package server handles the HTTP service component of the PDP.
package server

import (
	"context"
	"log/slog"
	"sync"
	"time"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller/adl"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config"
)

// decisionLogShutdownTimeout bounds how long Serve waits for queued decisions to flush on shutdown.
const decisionLogShutdownTimeout = 10 * time.Second

// NewService initializes the HTTP service (implemented with fiber & fasthttp).
func NewService(cfg *config.Config, logger *slog.Logger) *Services {
	s := &Services{cfg: cfg, logger: logger, chk: handle.NewChecks()}

	s.chk.SetHealth(true)
	s.chk.SetAlive(true)
	s.chk.SetReady(false) // wait for latest bundle first!

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

	s.main = fiber.New(logger, s.initMainRoutes, opts...)

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

	s.bundles = fiber.New(logger, s.initBundlesRoutes, opts...)

	// health service
	s.health = fiber.New(
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
// It blocks until all HTTP(S) services have stopped serving requests, via Shutdown,
// or an OS signal each service catches independently and only then flushes the authorization decision log.
func (s *Services) Serve() {
	wg := sync.WaitGroup{}
	wg.Add(3)

	go s.bundles.ServeWithWG(&wg)
	go s.main.ServeWithWG(&wg)
	go s.health.ServeWithWG(&wg)

	wg.Wait()

	s.flushDecisionLog()
}

// flushDecisionLog blocks until any authorization decisions queued for export by the batch span processor
// have been flushed, so a SIGTERM doesn't silently drop decisions the caller already received a response for.
func (s *Services) flushDecisionLog() {
	if s.decisionLog == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), decisionLogShutdownTimeout)
	defer cancel()

	if err := s.decisionLog.Shutdown(ctx); err != nil {
		s.logger.Error("failed to flush authorization decision log", "error", err)
	}
}

// Shutdown shuts down the HTTP(S) services.
func (s *Services) Shutdown() {
	s.health.Shutdown()
	s.main.Shutdown()
	s.bundles.Shutdown()
}

// Services contains the details of the HTTP(S) services.
type Services struct {
	ctx         context.Context
	logger      *slog.Logger
	cfg         *config.Config
	decisionLog *adl.ADL
	l           models.Language
	main        server.Service
	bundles     server.Service
	health      server.Service
	auth        *authHandler
	chk         *handle.Checks
}
