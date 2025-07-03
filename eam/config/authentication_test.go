package config

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestAuthentication_NewAuthenticator(t *testing.T) {
	t.Parallel()

	t.Run("new authenticator", func(t *testing.T) {
		a1 := &Authentication{Type: "bcrypt"}

		ctx := context.Background()

		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		c := cedar_embedded.NewController(
			pdp.WithContext(ctx),
			pdp.WithLogger(logger),
			pdp.WithPEP(pep.New(ctx, logger)),
			pdp.WithPIP(pip.New(ctx, logger)),
			pdp.WithPAP(pap.New(ctx, logger)),
		)

		a2, err := a1.NewAuthenticator(c)
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

		c := cedar_embedded.NewController(
			pdp.WithContext(ctx),
			pdp.WithLogger(logger),
			pdp.WithPEP(pep.New(ctx, logger)),
			pdp.WithPIP(pip.New(ctx, logger)),
			pdp.WithPAP(pap.New(ctx, logger)),
		)

		a1 := &Authentication{Type: "oopsie"}
		a2, err := a1.NewAuthenticator(c)
		require.NoError(t, err)
		require.Nil(t, a2)
	})
}
