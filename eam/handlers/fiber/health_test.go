package fiber

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
)

func TestHealthy(t *testing.T) {
	t.Parallel()

	t.Run("test healthy", func(t *testing.T) {
		chk := NewChecks()
		require.NotNil(t, chk)

		chk.SetHealth(true)

		srv := fiber.New()
		srv.Get("/healthz", chk.HealthZ)

		req := httptest.NewRequest("GET", "/healthz", nil)
		resp, err2 := srv.Test(req, 100)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, HealthVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		b, err3 := io.ReadAll(resp.Body)
		assert.NoError(t, err3)
		assert.EqualValues(t, server.ResponseBody[200], b)
	})
}

func TestNotHealthy(t *testing.T) {
	t.Parallel()

	t.Run("test not healthy", func(t *testing.T) {
		chk := NewChecks()
		require.NotNil(t, chk)

		srv := fiber.New()
		srv.Get("/healthz", chk.HealthZ)

		req := httptest.NewRequest("GET", "/healthz", nil)
		resp, err2 := srv.Test(req, 100)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, HealthVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
	})
}

func TestAlive(t *testing.T) {
	t.Parallel()

	t.Run("test alive", func(t *testing.T) {
		chk := NewChecks()
		require.NotNil(t, chk)

		chk.SetAlive(true)

		srv := fiber.New()
		srv.Get("/livez", chk.LiveZ)

		req := httptest.NewRequest("GET", "/livez", nil)
		resp, err2 := srv.Test(req, 100)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, HealthVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		b, err3 := io.ReadAll(resp.Body)
		assert.NoError(t, err3)
		assert.EqualValues(t, server.ResponseBody[200], b)
	})
}

func TestNotAlive(t *testing.T) {
	t.Parallel()

	t.Run("test not alive", func(t *testing.T) {
		chk := NewChecks()
		require.NotNil(t, chk)

		srv := fiber.New()
		srv.Get("/livez", chk.LiveZ)

		req := httptest.NewRequest("GET", "/livez", nil)
		resp, err2 := srv.Test(req, 100)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, HealthVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
	})
}

func TestReady(t *testing.T) {
	t.Parallel()

	t.Run("test ready", func(t *testing.T) {
		chk := NewChecks()
		require.NotNil(t, chk)

		chk.SetReady(true)

		srv := fiber.New()
		srv.Get("/healthz", chk.ReadyZ)

		req := httptest.NewRequest("GET", "/healthz", nil)
		resp, err2 := srv.Test(req, 100)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, HealthVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		b, err3 := io.ReadAll(resp.Body)
		assert.NoError(t, err3)
		assert.EqualValues(t, server.ResponseBody[200], b)
	})
}

func TestNotReady(t *testing.T) {
	t.Parallel()

	t.Run("test not ready", func(t *testing.T) {
		chk := NewChecks()
		require.NotNil(t, chk)

		srv := fiber.New()
		srv.Get("/readyz", chk.ReadyZ)

		req := httptest.NewRequest("GET", "/readyz", nil)
		resp, err2 := srv.Test(req, 100)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, HealthVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
	})
}
