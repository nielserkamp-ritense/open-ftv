package handlers

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/config"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestAuthHandler_FSC(t *testing.T) {
	in := `{"input":{"method":"POST","path":"/x/y"}}`
	out := `{"result":{"allowed":true,"status":{"reason":"ok"}}}`

	t.Run("fsc handler", func(t *testing.T) {
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

		auth := New(cfg, logger, nil)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/auth", auth.AuthFSC)

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

func TestAuthHandler_FSC_Fail1(t *testing.T) {
	in := `{"input":{"method":"GET","path":"/x/y","headers":{"doelbinding":["subsidies"]}}}`
	out := `{"result":{"allowed":false,"status":{"reason":"not authorized"}}}`

	t.Run("fsc handler fail (1)", func(t *testing.T) {
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

		auth := New(cfg, logger, nil)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/auth", auth.AuthFSC)

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

func TestAuthHandler_FSC_Fail2(t *testing.T) {
	in := `{"input":{}}`
	out := `{"message":"invalid method"}`

	t.Run("fsc handler fail (2)", func(t *testing.T) {
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

		auth := New(cfg, logger, nil)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/auth", auth.AuthFSC)

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

func TestAuthHandler_FSC_Fail3(t *testing.T) {
	in := `[]`
	out := `{"message":"invalid input data"}`

	t.Run("fsc handler fail (3)", func(t *testing.T) {
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

		auth := New(cfg, logger, nil)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/auth", auth.AuthFSC)

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

func TestAuthHandler_FSC_Fail4(t *testing.T) {
	in := `{"input":{"method":"POST"}}`
	out := `{"message":"invalid path"}`

	t.Run("fsc handler fail (4)", func(t *testing.T) {
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

		auth := New(cfg, logger, nil)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/auth", auth.AuthFSC)

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
