package fiber

import (
	"bytes"
	"context"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// TestBodyLimit pins the request body limit of a service built with New to Fiber's built-in
// default, as the READMEs promise. It fails when a Fiber upgrade changes the default or when
// someone sets BodyLimit again.
func TestBodyLimit(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{name: "at limit", size: fiber.DefaultBodyLimit},
		{name: "over limit", size: fiber.DefaultBodyLimit + 1, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))
			router := func(_ context.Context, svc *fiber.App) {
				svc.Post("/", func(c *fiber.Ctx) error {
					return c.SendStatus(fiber.StatusOK)
				})
			}

			s, ok := New(logger, router, server.WithDefaults()).(*service)
			require.True(t, ok)

			req := httptest.NewRequest(fiber.MethodPost, "/", bytes.NewReader(bytes.Repeat([]byte("a"), tc.size)))
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			assert.Equal(t, fiber.DefaultBodyLimit, s.svc.Config().BodyLimit)

			// fasthttp reports an oversized body on the in-memory test connection as an error
			// rather than as a 413 response.
			resp, err := s.svc.Test(req, -1)
			if tc.wantErr {
				require.ErrorContains(t, err, "body size exceeds the given limit")

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			defer resp.Body.Close()

			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		})
	}
}
