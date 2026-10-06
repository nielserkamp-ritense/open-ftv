package fiber

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestServe(t *testing.T) {
	t.Parallel()

	t.Run("serve", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		opts := []server.Option{
			server.WithDefaults(),
			server.WithHostPort("127.0.0.1", 20000),
		}

		s := New(logger, nil, opts...)
		require.NotNil(t, s)
		assert.NotNil(t, s.Context())
		assert.Equal(t, logger, s.Logger())

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
	t.Parallel()

	t.Run("error handler", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		opts := []server.Option{
			server.WithDefaults(),
			server.WithHostPort("127.0.0.1", 20002),
			server.WithRecovery(),
			server.WithSecurity(),
			server.WithCORS("*", "Cache-Control"),
		}

		s := New(logger, func(_ context.Context, svc *fiber.App) {
			svc.Get("/healthz", func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusOK)
			})
		}, opts...)

		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			s.Serve()
			wg.Done()
		}()

		go func() {
			time.Sleep(50 * time.Millisecond)

			resp, err := http.DefaultClient.Get("http://127.0.0.1:20002")
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

			time.Sleep(10 * time.Millisecond)

			s.Shutdown()
			wg.Done()
		}()

		wg.Wait()

		assert.GreaterOrEqual(t, h.Count(), 4)
	})
}

func TestRecovery(t *testing.T) {
	t.Parallel()

	t.Run("panic recovery", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		opts := []server.Option{
			server.WithDefaults(),
			server.WithHostPort("127.0.0.1", 20003),
			server.WithRecovery(),
		}

		router := func(_ context.Context, svc *fiber.App) {
			svc.Get("/healthz", func(req *fiber.Ctx) error {
				panic("oops")
			})
		}

		s := New(logger, router, opts...)

		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			s.Serve()
			wg.Done()
		}()

		go func() {
			time.Sleep(50 * time.Millisecond)

			resp, err := http.DefaultClient.Get("http://127.0.0.1:20003/healthz")
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

			time.Sleep(10 * time.Millisecond)

			s.Shutdown()
			wg.Done()
		}()

		wg.Wait()

		assert.GreaterOrEqual(t, h.Count(), 4)
	})
}

func TestInvalidHostPort(t *testing.T) {
	t.Parallel()

	t.Run("invalid host/port", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		opts := []server.Option{
			server.WithDefaults(),
			server.WithHostPort("what-is-it?", 1),
			server.WithRecovery(),
		}

		router := func(_ context.Context, svc *fiber.App) {
			svc.Get("/healthz", func(req *fiber.Ctx) error {
				return req.SendStatus(fiber.StatusOK)
			})
		}

		s := New(logger, router, opts...)

		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			s.Serve()
			wg.Done()
		}()

		go func() {
			time.Sleep(25 * time.Millisecond)
			wg.Done()
		}()

		wg.Wait()

		assert.GreaterOrEqual(t, h.Count(), 2)
	})
}

func TestForceAbort(t *testing.T) {
	t.Parallel()

	t.Run("force abort", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		opts := []server.Option{
			server.WithDefaults(),
			server.WithHostPort("0.0.0.0", 20005),
			server.WithRecovery(),
		}

		router := func(_ context.Context, svc *fiber.App) {
			svc.Get("/healthz", func(req *fiber.Ctx) error {
				return req.SendStatus(fiber.StatusOK)
			})
		}

		s := New(logger, router, opts...)

		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			s.Serve()
			wg.Done()
		}()

		go func() {
			time.Sleep(25 * time.Millisecond)
			s.Shutdown()

			// a second OS signal while stopping forces the stop.
			select {
			case s.(*service).intChan <- syscall.SIGQUIT:
			case <-time.After(time.Second):
			}

			wg.Done()
		}()

		wg.Wait()

		assert.GreaterOrEqual(t, h.Count(), 1)
	})
}

func TestShutdownTwiceStopsGracefully(t *testing.T) {
	t.Parallel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	s := New(slog.New(h), nil, server.WithDefaults(), server.WithHostPort("127.0.0.1", 20006))

	done := make(chan struct{})

	go func() {
		s.Serve()
		close(done)
	}()

	time.Sleep(25 * time.Millisecond)
	s.Shutdown()
	s.Shutdown()
	<-done

	assert.Contains(t, h.Log(), "http service stopped successfully")
	assert.NotContains(t, h.Log(), "aborted")
}
