package fiber

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server/fiber"
)

func TestHealth(t *testing.T) {
	t.Run("test health", func(t *testing.T) {
		srv := fiber.New()
		srv.Get("/healthz", HealthZ)

		req := httptest.NewRequest("GET", "/healthz", nil)
		resp, err2 := srv.Test(req, 5)

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
