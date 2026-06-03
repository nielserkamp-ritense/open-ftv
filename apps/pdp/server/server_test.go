package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config"
)

func TestServe(t *testing.T) {
	t.Parallel()

	t.Run("serve", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		cfg := &config.Config{
			ServerApp: config2.ServerApp{
				Server: config2.Server{
					Host:         "127.0.0.1",
					Port:         23000,
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 10 * time.Second,
					IdleTimeout:  300 * time.Second,
					MaxBody:      64536,
				},
			},
			PAP: config2.PAP{Language: "cedar"},
		}

		s := NewService(cfg, logger)

		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			s.Serve()
			wg.Done()
		}()

		go func() {
			time.Sleep(50 * time.Millisecond)
			s.Shutdown()
			wg.Done()
		}()

		wg.Wait()
	})
}

func TestServe_FailPDP(t *testing.T) {
	t.Parallel()

	t.Run("fail PDP", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		cfg := &config.Config{
			ServerApp: config2.ServerApp{
				Server: config2.Server{
					Host:         "127.0.0.1",
					Port:         23001,
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 10 * time.Second,
					IdleTimeout:  300 * time.Second,
					MaxBody:      64536,
				},
			},
		}

		defer func() {
			e := recover()
			require.NotNil(t, e)
		}()

		_ = NewService(cfg, logger)

		require.True(t, false) // should never trigger
	})
}

func TestErrorHandler(t *testing.T) {
	t.Parallel()

	t.Run("error handler", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		cfg := &config.Config{
			ServerApp: config2.ServerApp{
				Server: config2.Server{
					Host:         "127.0.0.1",
					Port:         23002,
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 10 * time.Second,
					IdleTimeout:  300 * time.Second,
					MaxBody:      64536,
				},
			},
			PAP: config2.PAP{Language: "cedar"},
		}

		s := NewService(cfg, logger)

		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			s.Serve()
			wg.Done()
		}()

		go func(cfg *config.Config) {
			time.Sleep(50 * time.Millisecond)

			resp, err := http.DefaultClient.Get(fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port))
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, 404, resp.StatusCode)

			time.Sleep(10 * time.Millisecond)

			s.Shutdown()
			wg.Done()
		}(cfg)

		wg.Wait()

		assert.GreaterOrEqual(t, h.Count(), 7)
	})
}
