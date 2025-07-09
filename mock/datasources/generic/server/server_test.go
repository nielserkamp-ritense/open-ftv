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
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/generic/config"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestServe(t *testing.T) {
	t.Run("serve", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		cfg := &config.Config{
			ServerApp: config2.ServerApp{
				Server: config2.Server{
					Host:         "127.0.0.1",
					Port:         20010,
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 10 * time.Second,
					IdleTimeout:  300 * time.Second,
					MaxBody:      256,
				},
			},
			DataPath: "../../../../testdata/dataspaces/fds",
		}

		s, err := NewService(cfg, logger)
		require.NoError(t, err)

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

		assert.GreaterOrEqual(t, h.Count(), 3)
	})
}

func TestErrorHandler(t *testing.T) {
	t.Run("error handler", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		cfg := &config.Config{
			ServerApp: config2.ServerApp{
				Server: config2.Server{
					Host:         "127.0.0.1",
					Port:         20011,
					ReadTimeout:  10 * time.Second,
					WriteTimeout: 10 * time.Second,
					IdleTimeout:  300 * time.Second,
					MaxBody:      256,
				},
			},
			DataPath: "../../../../testdata/dataspaces/fds",
		}

		s, err := NewService(cfg, logger)
		require.NoError(t, err)

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

		assert.GreaterOrEqual(t, 8, h.Count())
	})
}
