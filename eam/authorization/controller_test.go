package authorization

import (
	"context"
	"log/slog"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

// TestNewForController_FailClosedOnEmpty verifies that an empty policy store allows all
// requests by default (legacy fail-open) but denies them when failClosedOnEmpty is set.
func TestNewForController_FailClosedOnEmpty(t *testing.T) {
	t.Parallel()

	build := func(failClosed bool) RequestAuthorizer {
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
		az, err := NewForController(c, nil, false, failClosed)
		require.NoError(t, err)
		require.NotNil(t, az)
		return az
	}

	uid := uuid.New()
	u, _ := url.Parse("http://localhost/v1/policies")
	req := &Request{UID: &uid, URL: u, Method: "GET"}

	t.Run("fail open (default) allows when store empty", func(t *testing.T) {
		t.Parallel()

		resp, _, err := build(false).AuthorizeRequest(req)
		require.NoError(t, err)
		require.True(t, resp.Allowed)
	})

	t.Run("fail closed denies when store empty", func(t *testing.T) {
		t.Parallel()

		resp, _, err := build(true).AuthorizeRequest(req)
		require.NoError(t, err)
		require.False(t, resp.Allowed)
	})
}

func TestNewForController(t *testing.T) {
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

		a2, err := authentication.New(ctx, logger, "bcrypt", p1.GetEntity)
		require.NoError(t, err)
		require.NotNil(t, a2)

		a4, err2 := NewForController(c, a2, true, false)
		require.NoError(t, err2)
		require.NotNil(t, a4)
	})
}

func TestNewForController_Fail(t *testing.T) {
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

		a2, err2 := NewForController(c, nil, true, false)
		require.Error(t, err2)
		require.Nil(t, a2)
	})
}
