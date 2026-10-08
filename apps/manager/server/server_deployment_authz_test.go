//go:build integration

package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/bundles"
)

// Test_Deployment_RoleMatrix drives the SCRUM-16 matrix for deployments through the manager
// as deployed: secured mode, OIDC, the shipped policies seeded into postgres. Publishing is
// the `deploy` action of the author; admin and auditor read.
//
// The manager schedules its bootstrap deployment 15 s after start when nothing was published;
// this test publishes well within that window (about 2 s), so the bootstrap run cannot refuse
// the author's deployment as still active.
func Test_Deployment_RoleMatrix(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)
	cnf.PAP.Store = "../../../testdata/apps/manager/policies/cedar"
	cnf.Authorization.FailClosedOnEmpty = true

	ip := newIDP(t)
	cnf.OIDC = config2.OIDC{JWKSURL: ip.srv.URL, Issuer: "https://idp.test", Audience: "openftv"}

	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))
	app := NewExternal(cnf, lgr).GetMainService().GetFiberApp()

	author, auditor, admin := ip.token(t, "author-user", "author"), ip.token(t, "auditor-user", "auditor"), ip.token(t, "admin-user", "admin")

	t.Run("everyone reads the collection and its vocabulary", func(t *testing.T) {
		for name, tok := range map[string]string{"author": author, "auditor": auditor, "admin": admin} {
			res := call(t, app, http.MethodGet, "/v1/deployments", tok, nil)
			require.NoError(t, res.Body.Close())
			require.Equal(t, http.StatusOK, res.StatusCode, name)

			res = call(t, app, http.MethodGet, "/v1/statuses", tok, nil)
			require.NoError(t, res.Body.Close())
			require.Equal(t, http.StatusOK, res.StatusCode, name)

			res = call(t, app, http.MethodGet, "/v1/bundle-configurations", tok, nil)
			require.NoError(t, res.Body.Close())
			require.Equal(t, http.StatusOK, res.StatusCode, name)
		}
	})

	t.Run("no token is denied, not served", func(t *testing.T) {
		res := call(t, app, http.MethodGet, "/v1/deployments", "", nil)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusForbidden, res.StatusCode)
	})

	t.Run("a denied caller is answered 403 before a not-found and before a validation error", func(t *testing.T) {
		res := call(t, app, http.MethodGet, "/v1/deployment/1", "", nil)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusForbidden, res.StatusCode)

		res = call(t, app, http.MethodGet, "/v1/deployment/999", "", nil)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusForbidden, res.StatusCode)

		res = call(t, app, http.MethodGet, "/v1/deployment/last", "", nil)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusForbidden, res.StatusCode)

		res = call(t, app, http.MethodGet, "/v1/deployment/abc", "", nil)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusForbidden, res.StatusCode, "a malformed version is 403, not 400")

		res = callRaw(t, app, http.MethodPost, "/v1/deployment", auditor, "not json")
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusForbidden, res.StatusCode, "a malformed body is 403, not 400")
	})

	t.Run("a permitted caller gets the store's answer: not found, or a validation error", func(t *testing.T) {
		res := call(t, app, http.MethodGet, "/v1/deployment/last", auditor, nil)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusNotFound, res.StatusCode)

		res = call(t, app, http.MethodGet, "/v1/deployment/999", auditor, nil)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusNotFound, res.StatusCode)

		res = call(t, app, http.MethodGet, "/v1/deployment/abc", auditor, nil)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusBadRequest, res.StatusCode)

		res = callRaw(t, app, http.MethodPost, "/v1/deployment", author, "not json")
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

	t.Run("only the author publishes", func(t *testing.T) {
		body := oas.NewDeploymentBody{Title: "Publicatie"}

		res := call(t, app, http.MethodPost, "/v1/deployment", auditor, body)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusForbidden, res.StatusCode)

		res = call(t, app, http.MethodPost, "/v1/deployment", admin, body)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusForbidden, res.StatusCode)

		res = call(t, app, http.MethodPost, "/v1/deployment", author, body)
		require.NoError(t, res.Body.Close())
		require.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("everyone reads the published deployment", func(t *testing.T) {
		for name, tok := range map[string]string{"author": author, "auditor": auditor, "admin": admin} {
			res := call(t, app, http.MethodGet, "/v1/deployment/last", tok, nil)
			require.NoError(t, res.Body.Close())
			require.Equal(t, http.StatusOK, res.StatusCode, name)

			res = call(t, app, http.MethodGet, "/v1/deployment/1", tok, nil)
			require.NoError(t, res.Body.Close())
			require.Equal(t, http.StatusOK, res.StatusCode, name)
		}
	})
}

// callRaw sends body as-is, for requests the JSON encoder would refuse to produce.
func callRaw(t *testing.T, app *fiber.App, method, path, bearer, body string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}

	res, err := app.Test(req, 15000)
	require.NoError(t, err)

	return res
}
