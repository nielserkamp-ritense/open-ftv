package fiber

import (
	"bytes"
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
	authfiber "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	bundles2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/bundles"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestNewBundlesHandler(t *testing.T) {
	t.Parallel()

	t.Run("test new bundles handler", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""))
		require.NoError(t, err)
		require.NotNil(t, p1)

		manager := bundles.NewManager(ctx, logger, bundles.WithConfig("../../../testdata/unittest/bundles/test2", true))
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

		p1, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""))
		require.NoError(t, err)
		require.NotNil(t, p1)

		manager := bundles.NewManager(ctx, logger,
			bundles.WithConfig("../../../testdata/unittest/bundles/test2", true),
			bundles.WithPolicyHandler(p1),
		)
		require.NotNil(t, manager)

		auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
		require.NotNil(t, auth)

		ah := NewBundlesHandler(logger, p1, manager, auth)
		require.NotNil(t, ah)

		srv := fiber.New()
		srv.Get("/v1/statuses", ah.GetStatuses)

		req := httptest.NewRequestWithContext(ctx, fiber.MethodGet, "/v1/statuses", nil)
		resp, err2 := srv.Test(req)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, BundlesVersion, resp.Header.Get(HeaderVersion))

		b, err3 := io.ReadAll(resp.Body)
		require.NoError(t, err3)
		require.NotNil(t, b)

		var list bundles2.Statuses

		err = json.Unmarshal(b, &list)
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

		p1, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""))
		require.NoError(t, err)
		require.NotNil(t, p1)

		manager := bundles.NewManager(ctx, logger,
			bundles.WithConfig("../../../testdata/unittest/bundles/test2", true),
			bundles.WithPolicyHandler(p1),
		)
		require.NotNil(t, manager)

		auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
		require.NotNil(t, auth)

		ah := NewBundlesHandler(logger, p1, manager, auth)
		require.NotNil(t, ah)

		srv := fiber.New()
		srv.Get("/v1/compression-types", ah.GetCompressTypes)

		req := httptest.NewRequestWithContext(ctx, fiber.MethodGet, "/v1/compression-types", nil)
		resp, err2 := srv.Test(req)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, BundlesVersion, resp.Header.Get(HeaderVersion))

		b, err3 := io.ReadAll(resp.Body)
		require.NoError(t, err3)
		require.NotNil(t, b)

		var list bundles2.CompressTypes

		err = json.Unmarshal(b, &list)
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

		p1, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""))
		require.NoError(t, err)
		require.NotNil(t, p1)

		manager := bundles.NewManager(ctx, logger,
			bundles.WithConfig("../../../testdata/unittest/bundles/test2", true),
			bundles.WithPolicyHandler(p1),
		)
		require.NotNil(t, manager)

		auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
		require.NotNil(t, auth)

		ah := NewBundlesHandler(logger, p1, manager, auth)
		require.NotNil(t, ah)

		srv := fiber.New()
		srv.Get("/v1/configs", ah.GetConfigs)

		req := httptest.NewRequestWithContext(ctx, fiber.MethodGet, "/v1/configs", nil)
		resp, err2 := srv.Test(req)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, BundlesVersion, resp.Header.Get(HeaderVersion))

		b, err3 := io.ReadAll(resp.Body)
		require.NoError(t, err3)
		require.NotNil(t, b)

		var list bundles2.BundleConfigs

		err = json.Unmarshal(b, &list)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 3)
	})
}

func TestBundlesHandler_PostDeployment(t *testing.T) {
	t.Parallel()

	t.Run("post deployment", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""))
		require.NoError(t, err)
		require.NotNil(t, p1)

		manager := bundles.NewManager(ctx, logger,
			bundles.WithConfig("../../../testdata/unittest/bundles", false),
			bundles.WithPolicyHandler(p1),
		)
		require.NotNil(t, manager)

		auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
		require.NotNil(t, auth)

		ah := NewBundlesHandler(logger, p1, manager, auth)
		require.NotNil(t, ah)

		srv := fiber.New()
		srv.Post("/v1/deployment", ah.PostDeployment)

		req := httptest.NewRequestWithContext(ctx, fiber.MethodPost, "/v1/deployment", bytes.NewBufferString(`{"description":"yoyo"}`))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		resp, err2 := srv.Test(req)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, BundlesVersion, resp.Header.Get(HeaderVersion))

		b, err3 := io.ReadAll(resp.Body)
		require.NoError(t, err3)
		require.NotNil(t, b)

		var d bundles2.Deployment

		err = json.Unmarshal(b, &d)
		require.NoError(t, err)
		assert.Equal(t, 1, d.Version)
		assert.Equal(t, "yoyo", d.Description)
		assert.Equal(t, int(bundles.Creating), d.Status)
	})
}

func TestBundlesHandler_GetDeployment(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		deploy     bool
		key        string
		wantStatus int
	}{
		{name: "no deployments", key: "1", wantStatus: fiber.StatusNotFound},
		{name: "bad key", deploy: true, key: "x", wantStatus: fiber.StatusBadRequest},
		{name: "wrong key", deploy: true, key: "2", wantStatus: fiber.StatusNotFound},
		{name: "good key", deploy: true, key: "1", wantStatus: fiber.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p1, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""))
			require.NoError(t, err)
			require.NotNil(t, p1)

			manager := bundles.NewManager(ctx, logger,
				bundles.WithConfig("../../../testdata/unittest/bundles", false),
				bundles.WithPolicyHandler(p1),
			)
			require.NotNil(t, manager)

			auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
			require.NotNil(t, auth)

			ah := NewBundlesHandler(logger, p1, manager, auth)
			require.NotNil(t, ah)

			srv := fiber.New()
			srv.Get("/v1/deployment/:key", ah.GetDeployment)
			srv.Post("/v1/deployment", ah.PostDeployment)

			if tc.deploy {
				req := httptest.NewRequestWithContext(ctx, fiber.MethodPost, "/v1/deployment", bytes.NewBufferString(`{"description":"yoyo"}`))
				req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
				resp, err2 := srv.Test(req)

				require.NoError(t, err2)
				require.NotNil(t, resp)
				defer resp.Body.Close()

				require.Equal(t, fiber.StatusOK, resp.StatusCode)
				assert.Equal(t, BundlesVersion, resp.Header.Get(HeaderVersion))

				b, err3 := io.ReadAll(resp.Body)
				require.NoError(t, err3)
				require.NotNil(t, b)

				var d bundles2.Deployment
				err := json.Unmarshal(b, &d)
				require.NoError(t, err)
				assert.Equal(t, 1, d.Version)
				assert.Equal(t, "yoyo", d.Description)
				assert.Equal(t, int(bundles.Creating), d.Status)
			}

			req2 := httptest.NewRequestWithContext(ctx, fiber.MethodGet, "/v1/deployment/"+tc.key, nil)
			resp2, err4 := srv.Test(req2)

			require.NoError(t, err4)
			require.NotNil(t, resp2)
			defer resp2.Body.Close()

			require.Equal(t, tc.wantStatus, resp2.StatusCode)

			if tc.wantStatus == fiber.StatusOK {
				assert.Equal(t, BundlesVersion, resp2.Header.Get(HeaderVersion))

				b2, err5 := io.ReadAll(resp2.Body)
				require.NoError(t, err5)
				require.NotNil(t, b2)

				var d2 bundles2.Deployment
				err5 = json.Unmarshal(b2, &d2)
				require.NoError(t, err5)
				assert.Equal(t, 1, d2.Version)
				assert.Equal(t, "yoyo", d2.Description)
			}
		})
	}
}

// newDeploymentServer mounts the deployment routes as the manager does, behind Identify, with
// the given authorizer deciding.
func newDeploymentServer(t *testing.T, auth authorization.RequestAuthorizer) *fiber.App {
	t.Helper()

	ctx := context.Background()
	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))

	p1, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""))
	require.NoError(t, err)

	manager := bundles.NewManager(ctx, logger,
		bundles.WithConfig("../../../testdata/unittest/bundles/test2", true),
		bundles.WithPolicyHandler(p1),
	)

	ah := NewBundlesHandler(logger, p1, manager, auth)
	identify := authfiber.Identify(auth, logger)

	srv := fiber.New()
	srv.Get("/v1/deployments", identify, ah.GetDeployments).
		Get("/v1/deployment/last", identify, ah.GetLastDeployment).
		Get("/v1/deployment/:key", identify, ah.GetDeployment).
		Post("/v1/deployment", identify, ah.PostDeployment)

	return srv
}

func status(t *testing.T, srv *fiber.App, method, path, body string) int {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := srv.Test(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp.StatusCode
}

// A caller the PDP denies is answered 403 for every operation, before a not-found and before
// a validation error, so nothing about the store's state or the request's shape leaks.
func TestBundlesHandler_deniedCallerGets403BeforeAnythingElse(t *testing.T) {
	t.Parallel()

	// no PDP and no NoAuth: every decision is a denial; no credentials: the system principal.
	srv := newDeploymentServer(t, authorization.New(authorization.WithAuthenticator(authentication.NewDummy())))

	assert.Equal(t, fiber.StatusForbidden, status(t, srv, fiber.MethodGet, "/v1/deployments", ""))
	assert.Equal(t, fiber.StatusForbidden, status(t, srv, fiber.MethodGet, "/v1/deployment/last", ""))
	assert.Equal(t, fiber.StatusForbidden, status(t, srv, fiber.MethodGet, "/v1/deployment/42", ""), "a missing version is 403, not 404")
	assert.Equal(t, fiber.StatusForbidden, status(t, srv, fiber.MethodGet, "/v1/deployment/abc", ""), "a malformed version is 403, not 400")
	assert.Equal(t, fiber.StatusForbidden, status(t, srv, fiber.MethodPost, "/v1/deployment", `{"title":"x"}`))
	assert.Equal(t, fiber.StatusForbidden, status(t, srv, fiber.MethodPost, "/v1/deployment", `not json`), "a malformed body is 403, not 400")
}

// A permitted caller gets the store's answer: not found while nothing was published, a
// validation error for a malformed request, and the deployment once published.
func TestBundlesHandler_permittedCallerGetsTheStoreAnswer(t *testing.T) {
	t.Parallel()

	srv := newDeploymentServer(t, authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy())))

	assert.Equal(t, fiber.StatusOK, status(t, srv, fiber.MethodGet, "/v1/deployments", ""))
	assert.Equal(t, fiber.StatusNotFound, status(t, srv, fiber.MethodGet, "/v1/deployment/last", ""))
	assert.Equal(t, fiber.StatusNotFound, status(t, srv, fiber.MethodGet, "/v1/deployment/42", ""))
	assert.Equal(t, fiber.StatusBadRequest, status(t, srv, fiber.MethodGet, "/v1/deployment/abc", ""))
	assert.Equal(t, fiber.StatusBadRequest, status(t, srv, fiber.MethodPost, "/v1/deployment", `not json`))

	assert.Equal(t, fiber.StatusOK, status(t, srv, fiber.MethodPost, "/v1/deployment", `{"title":"Publicatie"}`))
	assert.Equal(t, fiber.StatusOK, status(t, srv, fiber.MethodGet, "/v1/deployment/last", ""))
	assert.Equal(t, fiber.StatusOK, status(t, srv, fiber.MethodGet, "/v1/deployment/1", ""))
}
