package server

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNewPIP_NoPersistence(t *testing.T) {
	t.Parallel()

	h := slog2.NewDummyHandler(slog.LevelWarn)
	logger := slog.New(h)

	s := &Services{
		ctx:    context.Background(),
		cfg:    &config.Config{},
		logger: logger,
	}

	p, err := s.newPIP()
	require.ErrorIs(t, err, pip.ErrNoPersistence)
	require.Nil(t, p)
}
