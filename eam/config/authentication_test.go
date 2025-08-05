package config

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestAuthentication_NewAuthenticator(t *testing.T) {
	t.Parallel()

	t.Run("new authenticator", func(t *testing.T) {
		ctx := context.Background()

		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		p2 := pip.New(ctx, logger)

		a1 := &Authentication{Type: "bcrypt"}
		a2, err := a1.NewAuthenticator(ctx, logger, p2.GetEntity)
		require.NoError(t, err)
		require.NotNil(t, a2)
	})
}

func TestAuthentication_NewAuthenticator_Unknown(t *testing.T) {
	t.Parallel()

	t.Run("new authenticator", func(t *testing.T) {
		ctx := context.Background()

		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		p2 := pip.New(ctx, logger)

		a1 := &Authentication{Type: "oopsie"}
		a2, err := a1.NewAuthenticator(ctx, logger, p2.GetEntity)
		require.NoError(t, err)
		require.Nil(t, a2)
	})
}
