package fiber

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestErrorHandler_FiberError(t *testing.T) {
	t.Parallel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	logger := slog.New(h)

	srv := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(logger)})
	srv.Get("/test", func(_ *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "bad input")
	})

	resp, err := srv.Test(httptest.NewRequest("GET", "/test", http.NoBody))
	require.NoError(t, err)

	defer resp.Body.Close()

	b, err2 := io.ReadAll(resp.Body)
	require.NoError(t, err2)

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.Contains(t, string(b), "bad input")
}

func TestErrorHandler_FiberErrorNoMessage(t *testing.T) {
	t.Parallel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	logger := slog.New(h)

	srv := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(logger)})
	srv.Get("/test", func(_ *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "")
	})

	resp, err := srv.Test(httptest.NewRequest("GET", "/test", http.NoBody))
	require.NoError(t, err)

	defer resp.Body.Close()

	b, err2 := io.ReadAll(resp.Body)
	require.NoError(t, err2)

	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	require.Contains(t, string(b), utils.StatusMessage(fiber.StatusNotFound))
}

func TestErrorHandler_GenericError(t *testing.T) {
	t.Parallel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	logger := slog.New(h)

	srv := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(logger)})
	srv.Get("/test", func(_ *fiber.Ctx) error {
		return errors.New("boom")
	})

	resp, err := srv.Test(httptest.NewRequest("GET", "/test", http.NoBody))
	require.NoError(t, err)

	defer resp.Body.Close()

	b, err2 := io.ReadAll(resp.Body)
	require.NoError(t, err2)

	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	require.Contains(t, string(b), "An unexpected error occurred.")
}
