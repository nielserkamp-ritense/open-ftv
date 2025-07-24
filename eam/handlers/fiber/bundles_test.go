package fiber

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	bundles2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/bundles"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNewBundlesHandler(t *testing.T) {
	t.Parallel()

	t.Run("test new bundles handler", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pap.New(ctx, logger)
		require.NotNil(t, p1)

		manager := bundles.NewManager("../../../testdata/unittest/bundles/test2", true, logger)
		require.NotNil(t, manager)

		ah := NewBundlesHandler(logger, p1, manager, nil)
		require.NotNil(t, ah)
	})
}

func TestBundlesHandler_GetStatuses(t *testing.T) {
	t.Parallel()

	t.Run("get statuses", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pap.New(ctx, logger)
		require.NotNil(t, p1)

		manager := bundles.NewManager("../../../testdata/unittest/bundles/test2", true, logger)
		require.NotNil(t, manager)

		auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
		require.NotNil(t, auth)

		ah := NewBundlesHandler(logger, p1, manager, auth)
		require.NotNil(t, ah)

		srv := fiber.New()
		srv.Get("/v1/statuses", ah.GetStatuses)

		req := httptest.NewRequestWithContext(ctx, fiber.MethodGet, "/v1/statuses", nil)
		resp, err2 := srv.Test(req, 100)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, BundlesVersion, resp.Header.Get(HeaderVersion))

		b, err3 := io.ReadAll(resp.Body)
		require.NoError(t, err3)
		require.NotNil(t, b)

		var list bundles2.Statuses
		err := json.Unmarshal(b, &list)
		require.NoError(t, err)
		assert.Equal(t, int(bundles.StatusCount), len(list))
	})
}

func TestBundlesHandler_GetCompressTypes(t *testing.T) {
	t.Parallel()

	t.Run("get compression types", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pap.New(ctx, logger)
		require.NotNil(t, p1)

		manager := bundles.NewManager("../../../testdata/unittest/bundles/test2", true, logger)
		require.NotNil(t, manager)

		auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
		require.NotNil(t, auth)

		ah := NewBundlesHandler(logger, p1, manager, auth)
		require.NotNil(t, ah)

		srv := fiber.New()
		srv.Get("/v1/compression-types", ah.GetCompressTypes)

		req := httptest.NewRequestWithContext(ctx, fiber.MethodGet, "/v1/compression-types", nil)
		resp, err2 := srv.Test(req, 100)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, BundlesVersion, resp.Header.Get(HeaderVersion))

		b, err3 := io.ReadAll(resp.Body)
		require.NoError(t, err3)
		require.NotNil(t, b)

		var list bundles2.CompressTypes
		err := json.Unmarshal(b, &list)
		require.NoError(t, err)
		assert.Equal(t, int(bundles.CompressCount), len(list))
	})
}

func TestBundlesHandler_GetConfigs(t *testing.T) {
	t.Parallel()

	t.Run("get configs", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pap.New(ctx, logger)
		require.NotNil(t, p1)

		manager := bundles.NewManager("../../../testdata/unittest/bundles/test2", true, logger)
		require.NotNil(t, manager)

		auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
		require.NotNil(t, auth)

		ah := NewBundlesHandler(logger, p1, manager, auth)
		require.NotNil(t, ah)

		srv := fiber.New()
		srv.Get("/v1/configs", ah.GetConfigs)

		req := httptest.NewRequestWithContext(ctx, fiber.MethodGet, "/v1/configs", nil)
		resp, err2 := srv.Test(req, 100)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, BundlesVersion, resp.Header.Get(HeaderVersion))

		b, err3 := io.ReadAll(resp.Body)
		require.NoError(t, err3)
		require.NotNil(t, b)

		var list bundles2.BundleConfigs
		err := json.Unmarshal(b, &list)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 3)
	})
}
