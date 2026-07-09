package fiber

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// failingAuthLogger simulates a synchronous durable-log (WAL) append failure.
type failingAuthLogger struct{}

func (failingAuthLogger) Log(context.Context, bool, *authlog.AuthRecord) error {
	return errors.New("write-ahead log append failed")
}

// TestAuthProcess_WALFailure_FailsClosed covers finding A7: when the synchronous
// authorization-log append fails, the (permit) decision must not be returned. In
// strict mode the handler responds 500; with strict mode off, the earlier
// best-effort behaviour (decision returned despite the log failure) is restored.
func TestAuthProcess_WALFailure_FailsClosed(t *testing.T) {
	// Not parallel: toggles the package-level ADLStrict.
	const in = `{"subject":{"type":"user","id":"alice"},"action":{"name":"read","properties":{"method":"GET"}},"resource":{"type":"service","id":"svc"}}`

	resp := &models.Response{Allowed: true, Message: "ok"}
	logger := slog.New(slog2.NewDummyHandler(slog.LevelError))

	newApp := func() *fiber.App {
		auth := NewAuthHandlerZEN(logger, failingAuthLogger{}, &fakeController{resp: resp})
		app := fiber.New()
		app.Post("/v1/authzen", auth.Authorize)
		return app
	}

	post := func(app *fiber.App) (*http.Response, string) {
		req, err := http.NewRequest(fiber.MethodPost, "/v1/authzen", bytes.NewReader([]byte(in)))
		require.NoError(t, err)
		httpResp, err := app.Test(req, -1)
		require.NoError(t, err)
		body, _ := io.ReadAll(httpResp.Body)
		_ = httpResp.Body.Close()
		return httpResp, string(body)
	}

	t.Run("strict (default): WAL failure -> 500, no decision leaked", func(t *testing.T) {
		require.True(t, ADLStrict, "strict mode must default on")
		httpResp, body := post(newApp())
		assert.Equal(t, fiber.StatusInternalServerError, httpResp.StatusCode)
		assert.NotContains(t, body, "\"decision\":true")
	})

	t.Run("non-strict: decision returned despite WAL failure", func(t *testing.T) {
		ADLStrict = false
		defer func() { ADLStrict = true }()
		httpResp, body := post(newApp())
		assert.Equal(t, fiber.StatusOK, httpResp.StatusCode)
		assert.Contains(t, body, "\"decision\":true")
	})
}
