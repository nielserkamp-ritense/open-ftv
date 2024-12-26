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

func TestAuthHandler_AuthZEN1(t *testing.T) {
	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"context":{"en":"ok"},"decision":true}`

	t.Run("authzen handler (1)", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		cfg := &config.Config{
			Host:           "127.0.0.1",
			Port:           20020,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			IdleTimeout:    300 * time.Second,
			MaxBody:        64536,
			PolicyLanguage: "cedar",
			PolicyStore:    "../../../../../testdata/unittest/cedar",
		}

		auth := New(nil, cfg, logger, nil)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzenzen", auth.AuthZEN)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/authzenzen", buf)
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

func TestAuthHandler_AuthZEN2(t *testing.T) {
	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"context":{"en":"ok"},"decision":true}`

	t.Run("authzen handler (2)", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		cfg := &config.Config{
			Host:           "127.0.0.1",
			Port:           20021,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			IdleTimeout:    300 * time.Second,
			MaxBody:        64536,
			PolicyLanguage: "opa",
			PolicyStore:    "../../../../../testdata/unittest/opa",
		}

		auth := New(nil, cfg, logger, nil)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzenzen", auth.AuthZEN)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/authzenzen", buf)
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

func TestAuthHandler_AuthZEN_Fail1(t *testing.T) {
	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"GET"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"context":{"en":"not authorized"},"decision":false}`

	t.Run("authzen handler fail (1)", func(t *testing.T) {
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

		auth := New(nil, cfg, logger, nil)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.AuthZEN)

		cfg.PolicyLanguage = "" // this forces the bad config!

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/authzen", buf)
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

func TestAuthHandler_AuthZEN_Fail2(t *testing.T) {
	in := `{"subject":{"type":"","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"message":"invalid subject"}`

	t.Run("authzen handler fail (2)", func(t *testing.T) {
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

		auth := New(nil, cfg, logger, nil)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.AuthZEN)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/authzen", buf)
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

func TestAuthHandler_AuthZEN_Fail3(t *testing.T) {
	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"message":"invalid action"}`

	t.Run("authzen handler fail (3)", func(t *testing.T) {
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

		auth := New(nil, cfg, logger, nil)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.AuthZEN)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/authzen", buf)
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

func TestAuthHandler_AuthZEN_Fail4(t *testing.T) {
	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"POST"}},"resource":{"type":"service"}}`
	out := `{"message":"invalid resource"}`

	t.Run("authzen handler fail (4)", func(t *testing.T) {
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

		auth := New(nil, cfg, logger, nil)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.AuthZEN)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/authzen", buf)
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
