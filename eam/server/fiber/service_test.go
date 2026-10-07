package fiber

import (
	"context"
	"log/slog"
	"net"
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
			assert.Error(t, s.Serve())
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

	h := slog2.NewDummyHandler(slog.LevelInfo)

	entered := make(chan struct{})
	release := make(chan struct{})

	router := func(_ context.Context, svc *fiber.App) {
		svc.Get("/slow", func(req *fiber.Ctx) error {
			close(entered)
			<-release

			return req.SendStatus(fiber.StatusOK)
		})
	}

	s := New(slog.New(h), router, server.WithDefaults(), server.WithHostPort("127.0.0.1", 20005))

	done := make(chan error)

	go func() {
		done <- s.Serve()
	}()

	time.Sleep(25 * time.Millisecond)

	go func() {
		resp, err := http.DefaultClient.Get("http://127.0.0.1:20005/slow")
		if err == nil {
			_ = resp.Body.Close()
		}
	}()

	// the request still being answered keeps the graceful stop waiting.
	<-entered
	s.Shutdown()

	// a second OS signal while stopping forces the stop.
	s.(*service).intChan <- syscall.SIGQUIT

	require.NoError(t, <-done)
	close(release)

	assert.Contains(t, h.Log(), "http service aborted")
	assert.NotContains(t, h.Log(), "stopped successfully")
}

func TestShutdownBeforeServe(t *testing.T) {
	t.Parallel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	s := New(slog.New(h), nil, server.WithDefaults(), server.WithHostPort("127.0.0.1", 20007))

	s.Shutdown()
	require.NoError(t, s.Serve())

	// the port is free again: the server didn't start listening after all.
	ln, err := net.Listen("tcp", "127.0.0.1:20007")
	require.NoError(t, err)
	require.NoError(t, ln.Close())
}

func TestServePortTaken(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp4", "127.0.0.1:20008")
	require.NoError(t, err)

	t.Cleanup(func() { _ = ln.Close() })

	h := slog2.NewDummyHandler(slog.LevelInfo)
	s := New(slog.New(h), nil, server.WithDefaults(), server.WithHostPort("127.0.0.1", 20008))

	require.Error(t, s.Serve())
	assert.Contains(t, h.Log(), "failed to start http service")
}

func TestShutdownTwiceStopsGracefully(t *testing.T) {
	t.Parallel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	s := New(slog.New(h), nil, server.WithDefaults(), server.WithHostPort("127.0.0.1", 20006))

	done := make(chan struct{})

	go func() {
		assert.NoError(t, s.Serve())
		close(done)
	}()

	time.Sleep(25 * time.Millisecond)
	s.Shutdown()
	s.Shutdown()
	<-done

	assert.Contains(t, h.Log(), "http service stopped successfully")
	assert.NotContains(t, h.Log(), "aborted")
}
