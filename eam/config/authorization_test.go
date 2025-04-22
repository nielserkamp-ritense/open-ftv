package config

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pdp/cedar-embedded"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestAuthorization_NewAuthorizer(t *testing.T) {
	t.Parallel()

	t.Run("new authorization", func(t *testing.T) {
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

		a1 := &Authentication{Type: "bcrypt"}
		a2, err := a1.NewAuthenticator(c)
		require.NoError(t, err)
		require.NotNil(t, a2)

		a3 := &Authorization{Authenticate: true}
		a4, err2 := a3.NewAuthorizer(c, a2)
		require.NoError(t, err2)
		require.NotNil(t, a4)
	})
}

func TestAuthorization_NewAuthorizer_Fail(t *testing.T) {
	t.Parallel()

	t.Run("new authorization fail", func(t *testing.T) {
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

		a1 := &Authorization{Authenticate: true}
		a2, err2 := a1.NewAuthorizer(c, nil)
		require.Error(t, err2)
		require.Nil(t, a2)
	})
}
