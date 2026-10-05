//go:build integration

package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Test_LanguageAndDecision_RoleMatrix drives the matrix for the policy languages and the
// Authorization Decision Log through the manager as deployed: secured mode, OIDC, the shipped
// policies seeded into postgres. These were the last two routes on the HTTP-shaped model; with
// its rules gone from the seed directory, every role still reads both (SCRUM-146).
func Test_LanguageAndDecision_RoleMatrix(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)
	cnf.PAP.Store = "../../../testdata/apps/manager/policies/cedar"
	cnf.PIP.Store = "../../../testdata/apps/manager/data/entities"
	cnf.Authorization.FailClosedOnEmpty = true

	ip := newIDP(t)
	cnf.OIDC = config2.OIDC{JWKSURL: ip.srv.URL, Issuer: "https://idp.test", Audience: "openftv"}

	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))
	app := NewExternal(cnf, lgr).GetMainService().GetFiberApp()

	everyone := map[string]string{
		"author":  ip.token(t, "author-user", "author"),
		"auditor": ip.token(t, "auditor-user", "auditor"),
		"admin":   ip.token(t, "admin-user", "admin"),
	}

	t.Run("everyone reads the languages and the log", func(t *testing.T) {
		for name, tok := range everyone {
			require.Equal(t, http.StatusOK, call(t, app, http.MethodGet, "/v1/languages", tok, nil).StatusCode, name)
			require.Equal(t, http.StatusOK, call(t, app, http.MethodGet, "/v1/adl/entries?recent=1h", tok, nil).StatusCode, name)
		}
	})

	t.Run("no token is denied, not served", func(t *testing.T) {
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodGet, "/v1/languages", "", nil).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodGet, "/v1/adl/entries?recent=1h", "", nil).StatusCode)
	})

	t.Run("a role the examples do not know reads neither", func(t *testing.T) {
		user := ip.token(t, "plain-user", "user")
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodGet, "/v1/languages", user, nil).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodGet, "/v1/adl/entries?recent=1h", user, nil).StatusCode)
	})

	t.Run("an API key reads the languages, not the log", func(t *testing.T) {
		key := func(path string) int {
			req := httptest.NewRequest(http.MethodGet, path, http.NoBody)
			req.Header.Set(models.HeaderAPIKey, "1234567890abcdef") // system1.yaml

			res, err := app.Test(req, 15000)
			require.NoError(t, err)

			defer res.Body.Close()

			return res.StatusCode
		}

		require.Equal(t, http.StatusOK, key("/v1/languages"))
		require.Equal(t, http.StatusForbidden, key("/v1/adl/entries?recent=1h"))
	})

	t.Run("malformed criteria are a 400 for a reader and a 403 for anyone else", func(t *testing.T) {
		require.Equal(t, http.StatusBadRequest, call(t, app, http.MethodGet, "/v1/adl/entries?recent=soon", everyone["auditor"], nil).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodGet, "/v1/adl/entries?recent=soon", "", nil).StatusCode)
	})
}
