package authentication

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("new authenticator", func(t *testing.T) {
		ctx := context.Background()

		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		p2, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""))
		require.NoError(t, err)

		a2, err := New(ctx, logger, "bcrypt", p2.GetEntity)
		require.NoError(t, err)
		require.NotNil(t, a2)
	})
}

func TestNew_Unknown(t *testing.T) {
	t.Parallel()

	t.Run("new authenticator", func(t *testing.T) {
		ctx := context.Background()

		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		p2, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""))
		require.NoError(t, err)

		a2, err := New(ctx, logger, "oopsie", p2.GetEntity)
		require.NoError(t, err)
		require.Nil(t, a2)
	})
}
