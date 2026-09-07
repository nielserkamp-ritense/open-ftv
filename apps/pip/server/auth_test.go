package server

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pip/config"
	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNew(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{PAP: config2.PAP{Language: "CEDAR"}}

	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))

	s := &Services{ctx: context.Background(), logger: logger, cfg: cfg}
	s.l = models.LanguageFromString(cfg.Language)

	auth := s.newAuth()
	require.NotNil(t, auth)
	require.NotNil(t, auth.Controller())
	require.NotNil(t, auth.Controller().GetPAP())
}
