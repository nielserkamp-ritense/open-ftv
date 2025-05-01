package fiber

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestSendMessageResponse(t *testing.T) {
	t.Parallel()

	t.Run("send message response", func(t *testing.T) {
		t.Parallel()

		srv := fiber.New()
		srv.Get("/test", func(req *fiber.Ctx) error {
			return SendMessageResponse(req, fiber.StatusBadRequest, "I'm not here")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		resp, err2 := srv.Test(req, 100)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		b, err3 := io.ReadAll(resp.Body)
		require.NoError(t, err3)
		require.NotNil(t, b)

		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		require.True(t, strings.Contains(string(b), "I'm not here"))
	})
}
