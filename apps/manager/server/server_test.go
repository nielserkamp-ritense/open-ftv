//go:build integration

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

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config"
	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestServe(t *testing.T) {
	t.Parallel()

	t.Run("serve", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		cfg := newManagerConfigWithPostgres(t)
		cfg.Server = config2.Server{
			Host:         "127.0.0.1",
			Port:         21000,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  300 * time.Second,
		}

		s := NewExternal(cfg, logger)

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

func TestErrorHandler(t *testing.T) {
	t.Parallel()

	t.Run("error handler", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		cfg := newManagerConfigWithPostgres(t)
		cfg.Server = config2.Server{
			Host:         "127.0.0.1",
			Port:         21002,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  300 * time.Second,
		}

		s := NewExternal(cfg, logger)

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

			defer resp.Body.Close()

			assert.Equal(t, 404, resp.StatusCode)

			time.Sleep(10 * time.Millisecond)

			s.Shutdown()
			wg.Done()
		}(cfg)

		wg.Wait()

		assert.GreaterOrEqual(t, h.Count(), 9)
	})
}
