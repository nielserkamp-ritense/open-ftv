package server

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config"
	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNewPAP_NoPersistence(t *testing.T) {
	t.Parallel()

	h := slog2.NewDummyHandler(slog.LevelWarn)
	logger := slog.New(h)

	s := &Services{
		ctx:    context.Background(),
		cfg:    &config.Config{PAP: config2.PAP{Language: "cedar"}},
		l:      models.LanguageFromString("cedar"),
		logger: logger,
	}

	p, err := s.newPAP()
	require.ErrorIs(t, err, pap.ErrNoPersistence)
	require.Nil(t, p)
}
