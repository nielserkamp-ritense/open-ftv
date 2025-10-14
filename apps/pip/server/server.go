// Package server handles the HTTP(S) services of the PIP.
package server

import (
	"context"
	"log/slog"
	"sync"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pip/config"
	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
)

// NewServices initializes the HTTP(S) services (implemented with fiber & fasthttp).
func NewServices(cfg *config.Config, logger *slog.Logger) *Services {
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

	s.main = fiber.New(logger, s.initMainRoutes, opts...)

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
func (s *Services) Serve() {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go s.health.ServeWithWG(&wg)
	go s.main.ServeWithWG(&wg)

	wg.Wait()
}

// Shutdown shuts down the HTTP(S) services.
func (s *Services) Shutdown() {
	s.main.Shutdown()
	s.health.Shutdown()
}

// Services contains the details of the HTTP(S) services.
type Services struct {
	ctx    context.Context
	logger *slog.Logger
	cfg    *config.Config
	l      models.Language
	main   server.Service
	health server.Service
	auth   AuthHandler
	pip    *pip.PIP
	chk    *handle.Checks
}
