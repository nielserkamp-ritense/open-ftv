package server

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config"
)

func TestNew(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{PAP: config2.PAP{Language: "CEDAR"}}

	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))

	s := &Services{ctx: context.Background(), cfg: cfg, logger: logger, l: models.LanguageFromString(cfg.Language)}

	auth := s.newAuth("")
	require.NotNil(t, auth)
	require.NotNil(t, auth.controller)

	srv := fiber.New()
	srv.Get("/zen", auth.zen.Evaluation)

	req := httptest.NewRequest("GET", "/zen", http.NoBody)
	resp, err := srv.Test(req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
