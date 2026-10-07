package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"sync"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/laadpalen-applicatie/config"
)

type Services struct {
	assets     fs.FS
	runtimeEnv []byte
	main       server.Service
	health     server.Service
	chk        *handle.Checks
}

func NewService(cfg *config.Config, logger *slog.Logger, assets fs.FS, runtimeEnv map[string]string) (*Services, error) {
	if _, err := fs.Stat(assets, indexFile); err != nil {
		return nil, fmt.Errorf("the static assets have no %s, run `npm run build` before compiling: %w", indexFile, err)
	}

	s := &Services{
		assets:     assets,
		runtimeEnv: runtimeEnvScript(runtimeEnv),
		chk:        handle.NewChecks(),
	}

	s.chk.SetHealth(true)
	s.chk.SetAlive(true)
	s.chk.SetReady(true)

	s.main = fiber.New(
		logger,
		s.initMainRoutes,
		server.WithDefaults(),
		server.WithHostPort(cfg.Host, cfg.Port),
		server.WithTimeouts(cfg.ReadTimeout, cfg.WriteTimeout, cfg.IdleTimeout),
		server.WithRecovery(),
	)

	s.health = fiber.New(
		logger,
		s.initHealthRoutes,
		server.WithDefaults(),
		server.WithHostPort(cfg.HealthHost, cfg.HealthPort),
		server.WithSvcName("health"),
		server.WithTimeouts(cfg.HealthRead, cfg.HealthWrite, cfg.HealthIdle),
		server.WithRecovery(),
	)

	return s, nil
}

func (s *Services) Serve() error {
	services := []server.Service{s.main, s.health}

	errs := make([]error, len(services))
	wg := sync.WaitGroup{}

	// a service stops on an OS signal or when it fails to start; either way, the others stop too.
	for i, svc := range services {
		wg.Go(func() {
			errs[i] = svc.Serve()

			for _, other := range services {
				other.Shutdown()
			}
		})
	}

	wg.Wait()

	return errors.Join(errs...)
}

// runtimeEnvScript renders the JavaScript that exposes the runtime environment to the browser app.
// encoding/json escapes <, > and &, so values cannot break out of the script.
func runtimeEnvScript(env map[string]string) []byte {
	// encoding a map of strings cannot fail.
	values, _ := json.Marshal(env)

	return []byte("window.__ENV__ = " + string(values) + ";\n")
}
