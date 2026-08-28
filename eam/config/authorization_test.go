package config

import (
	"context"
	"log/slog"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	authorization2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

// TestAuthorization_FailClosedOnEmpty verifies that an empty policy store allows all
// requests by default (legacy fail-open) but denies them when FailClosedOnEmpty is set.
func TestAuthorization_FailClosedOnEmpty(t *testing.T) {
	t.Parallel()

	build := func(failClosed bool) authorization2.Authorizer {
		ctx := context.Background()
		logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))
		p1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""))
		require.NoError(t, err)

		ap1, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), "")) // empty store
		require.NoError(t, err)

		c := cedar_embedded.NewController(
			pdp.WithContext(ctx),
			pdp.WithLogger(logger),
			pdp.WithPEP(pep.New(ctx, logger)),
			pdp.WithPIP(p1),
			pdp.WithPAP(ap1),
		)
		a := &Authorization{FailClosedOnEmpty: failClosed}
		az, err := a.NewAuthorizer(c, nil)
		require.NoError(t, err)
		require.NotNil(t, az)
		return az
	}

	uid := uuid.New()
	u, _ := url.Parse("http://localhost/v1/policies")
	req := &authorization2.Request{UID: &uid, URL: u, Method: "GET"}

	t.Run("fail open (default) allows when store empty", func(t *testing.T) {
		t.Parallel()

		resp, _, err := build(false).Authorize(req)
		require.NoError(t, err)
		require.True(t, resp.Allowed)
	})

	t.Run("fail closed denies when store empty", func(t *testing.T) {
		t.Parallel()

		resp, _, err := build(true).Authorize(req)
		require.NoError(t, err)
		require.False(t, resp.Allowed)
	})
}

func TestAuthorization_NewAuthorizer(t *testing.T) {
	t.Parallel()

	t.Run("new authorization", func(t *testing.T) {
		ctx := context.Background()

		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		p1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""))
		require.NoError(t, err)

		ap1, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""))
		require.NoError(t, err)

		c := cedar_embedded.NewController(
			pdp.WithContext(ctx),
			pdp.WithLogger(logger),
			pdp.WithPEP(pep.New(ctx, logger)),
			pdp.WithPIP(p1),
			pdp.WithPAP(ap1),
		)

		a1 := &Authentication{Type: "bcrypt"}
		a2, err := a1.NewAuthenticator(ctx, logger, p1.GetEntity)
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

		p1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""))
		require.NoError(t, err)

		ap1, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""))
		require.NoError(t, err)

		c := cedar_embedded.NewController(
			pdp.WithContext(ctx),
			pdp.WithLogger(logger),
			pdp.WithPEP(pep.New(ctx, logger)),
			pdp.WithPIP(p1),
			pdp.WithPAP(ap1),
		)

		a1 := &Authorization{Authenticate: true}
		a2, err2 := a1.NewAuthorizer(c, nil)
		require.Error(t, err2)
		require.Nil(t, a2)
	})
}
