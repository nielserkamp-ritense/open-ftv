package handlers

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/config"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestAuthHandler_RunOK(t *testing.T) {
	in := `{"input":{"method":"POST","path":"/x/y"}}`
	out := `{"result":{"allowed":true}}`

	t.Run("auth handler", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		cfg := &config.Config{
			Host:           "127.0.0.1",
			Port:           20010,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			IdleTimeout:    300 * time.Second,
			MaxBody:        64536,
			PolicyLanguage: "cedar",
			PolicyStore:    "../../../../../testdata/unittest/cedar",
		}

		c, err := newController(cfg, logger, nil)
		require.NoError(t, err)
		require.NotNil(t, c)

		auth := &authHandler{cfg: cfg, logger: logger, controller: c}

		app := fiber.New()
		app.Post("/v1/auth", auth.run)

		cfg.PolicyLanguage = "" // this forces the bad config!

		wg := &sync.WaitGroup{}
		wg.Add(2)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/auth", buf)
		require.NoError(t, err2)
		require.NotNil(t, req)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
		assert.Equal(t, 8, h.Count())
	})
}

func TestAuthHandler_RunFail1(t *testing.T) {
	in := `{"input":{"method":"GET","path":"/x/y"}}`
	out := `{"result":{"allowed":false,"status":{"reason":"not authorized"}}}`

	t.Run("auth handler", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		cfg := &config.Config{
			Host:           "127.0.0.1",
			Port:           20010,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			IdleTimeout:    300 * time.Second,
			MaxBody:        64536,
			PolicyLanguage: "opa",
			PolicyStore:    "../../../../../testdata/unittest/opa",
		}

		c, err := newController(cfg, logger, nil)
		require.NoError(t, err)
		require.NotNil(t, c)

		auth := &authHandler{cfg: cfg, logger: logger, controller: c}

		app := fiber.New()
		app.Post("/v1/auth", auth.run)

		cfg.PolicyLanguage = "" // this forces the bad config!

		wg := &sync.WaitGroup{}
		wg.Add(2)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/auth", buf)
		require.NoError(t, err2)
		require.NotNil(t, req)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
		assert.Equal(t, 9, h.Count())
	})
}

func TestAuthHandler_RunFail2(t *testing.T) {
	in := `{"input":{}}`
	out := `{"message":"invalid method"}`

	t.Run("auth handler", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		cfg := &config.Config{
			Host:           "127.0.0.1",
			Port:           20010,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			IdleTimeout:    300 * time.Second,
			MaxBody:        64536,
			PolicyLanguage: "cerbos",
			PolicyStore:    "../../../../../testdata/unittest/cerbos",
		}

		c, err := newController(cfg, logger, nil)
		require.NoError(t, err)
		require.NotNil(t, c)

		auth := &authHandler{cfg: cfg, logger: logger, controller: c}

		app := fiber.New()
		app.Post("/v1/auth", auth.run)

		cfg.PolicyLanguage = "" // this forces the bad config!

		wg := &sync.WaitGroup{}
		wg.Add(2)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/auth", buf)
		require.NoError(t, err2)
		require.NotNil(t, req)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		defer resp.Body.Close()

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
		assert.Equal(t, 5, h.Count())
	})
}

func TestAuthHandler_RunFail3(t *testing.T) {
	in := `[]`
	out := `{"message":"invalid input data"}`

	t.Run("auth handler", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		cfg := &config.Config{
			Host:           "127.0.0.1",
			Port:           20010,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			IdleTimeout:    300 * time.Second,
			MaxBody:        64536,
			PolicyLanguage: "cerbos",
			PolicyStore:    "../../../../../testdata/unittest/cerbos",
		}

		c, err := newController(cfg, logger, nil)
		require.NoError(t, err)
		require.NotNil(t, c)

		auth := &authHandler{cfg: cfg, logger: logger, controller: c}

		app := fiber.New()
		app.Post("/v1/auth", auth.run)

		cfg.PolicyLanguage = "" // this forces the bad config!

		wg := &sync.WaitGroup{}
		wg.Add(2)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/auth", buf)
		require.NoError(t, err2)
		require.NotNil(t, req)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		defer resp.Body.Close()

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
		assert.Equal(t, 5, h.Count())
	})
}

func TestAuthHandler_RunFail4(t *testing.T) {
	in := `{"input":{"method":"POST"}}`
	out := `{"message":"invalid path"}`

	t.Run("auth handler", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		cfg := &config.Config{
			Host:           "127.0.0.1",
			Port:           20010,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			IdleTimeout:    300 * time.Second,
			MaxBody:        64536,
			PolicyLanguage: "cerbos",
			PolicyStore:    "../../../../../testdata/unittest/cerbos",
		}

		c, err := newController(cfg, logger, nil)
		require.NoError(t, err)
		require.NotNil(t, c)

		auth := &authHandler{cfg: cfg, logger: logger, controller: c}

		app := fiber.New()
		app.Post("/v1/auth", auth.run)

		cfg.PolicyLanguage = "" // this forces the bad config!

		wg := &sync.WaitGroup{}
		wg.Add(2)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/auth", buf)
		require.NoError(t, err2)
		require.NotNil(t, req)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		defer resp.Body.Close()

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
		assert.Equal(t, 5, h.Count())
	})
}
