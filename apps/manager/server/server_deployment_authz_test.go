//go:build integration

package server

import (
	"log/slog"
	"net/http"
	"os"
	"testing"

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
			require.Equal(t, http.StatusOK, call(t, app, http.MethodGet, "/v1/deployments", tok, nil).StatusCode, name)
			require.Equal(t, http.StatusOK, call(t, app, http.MethodGet, "/v1/statuses", tok, nil).StatusCode, name)
			require.Equal(t, http.StatusOK, call(t, app, http.MethodGet, "/v1/bundle-configurations", tok, nil).StatusCode, name)
		}
	})

	t.Run("no token is denied, not served", func(t *testing.T) {
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodGet, "/v1/deployments", "", nil).StatusCode)
	})

	t.Run("a denied caller cannot tell a missing version from an existing one", func(t *testing.T) {
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodGet, "/v1/deployment/1", "", nil).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodGet, "/v1/deployment/999", "", nil).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodGet, "/v1/deployment/last", "", nil).StatusCode)
	})

	t.Run("nothing published yet is not found for a permitted reader", func(t *testing.T) {
		require.Equal(t, http.StatusNotFound, call(t, app, http.MethodGet, "/v1/deployment/last", auditor, nil).StatusCode)
		require.Equal(t, http.StatusNotFound, call(t, app, http.MethodGet, "/v1/deployment/999", auditor, nil).StatusCode)
	})

	t.Run("only the author publishes", func(t *testing.T) {
		body := oas.NewDeploymentBody{Title: "Publicatie"}

		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPost, "/v1/deployment", auditor, body).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPost, "/v1/deployment", admin, body).StatusCode)
		require.Equal(t, http.StatusOK, call(t, app, http.MethodPost, "/v1/deployment", author, body).StatusCode)
	})

	t.Run("everyone reads the published deployment", func(t *testing.T) {
		for name, tok := range map[string]string{"author": author, "auditor": auditor, "admin": admin} {
			require.Equal(t, http.StatusOK, call(t, app, http.MethodGet, "/v1/deployment/last", tok, nil).StatusCode, name)
			require.Equal(t, http.StatusOK, call(t, app, http.MethodGet, "/v1/deployment/1", tok, nil).StatusCode, name)
		}
	})
}
