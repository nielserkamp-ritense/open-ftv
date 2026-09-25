package server

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pap/config"
	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNew(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{PAP: config2.PAP{Language: "CEDAR"}}

	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))

	s := &Services{ctx: context.Background(), logger: logger, cfg: cfg}
	s.l = models.LanguageFromString(cfg.Language)

	ap, err := pap.NewSelfAuthz(s.ctx, s.logger, s.cfg.Language, s.cfg.PAP.Store, s.cfg.PAP.StoreRecurse)
	require.NoError(t, err)

	s.pap = ap

	auth := s.newAuth()
	require.NotNil(t, auth)
	require.NotNil(t, auth.Controller())
	require.Same(t, ap, auth.Controller().GetPAP())
}
