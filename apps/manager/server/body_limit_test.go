package server

import (
	"bytes"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"

	eamconfig "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	serverfiber "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
)

func managerMaxBody(t *testing.T) int {
	t.Helper()

	cfg := &eamconfig.Server{}
	err := config.LoadConfig(cfg,
		config.NoFlags(),
		config.NoHelpOnError(),
		config.AppName("test"),
		// Unused prefix so a developer's MANAGER_MAX_BODY_SIZE does not override the app default.
		config.EnvironmentPrefix("TEST_MANAGER_BODY_UNUSED_"),
	)
	require.NoError(t, err)
	require.Greater(t, cfg.MaxBody, 0)

	return cfg.MaxBody
}

func TestMaxBodyRejectsOversizedRequest(t *testing.T) {
	t.Parallel()

	maxBody := managerMaxBody(t)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := fiber.New(fiber.Config{
		BodyLimit:    maxBody,
		ErrorHandler: serverfiber.ErrorHandler(logger),
	})
	app.Post("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	oversized := bytes.Repeat([]byte("a"), maxBody+1)
	req := httptest.NewRequest(fiber.MethodPost, "/", bytes.NewReader(oversized))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req, -1)
	if err != nil {
		// Fiber/fasthttp may surface BodyLimit as a test error instead of a response.
		assert.Contains(t, err.Error(), "body size exceeds the given limit")
		return
	}

	require.NotNil(t, resp)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestMaxBodyAcceptsRequestAtLimit(t *testing.T) {
	t.Parallel()

	maxBody := managerMaxBody(t)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := fiber.New(fiber.Config{
		BodyLimit:    maxBody,
		ErrorHandler: serverfiber.ErrorHandler(logger),
	})
	app.Post("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	body := bytes.Repeat([]byte("a"), maxBody)
	req := httptest.NewRequest(fiber.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NotNil(t, resp)

	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
