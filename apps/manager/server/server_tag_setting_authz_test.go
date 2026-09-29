//go:build integration

package server

import (
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
	settings "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/settings"
)

// Test_TagAndSetting_RoleMatrix drives the SCRUM-16 matrix for beslispunten and settings
// through the manager as deployed: secured mode, OIDC, the shipped policies seeded into
// postgres. Admin creates, updates and deletes; every role reads; author and auditor change
// nothing.
func Test_TagAndSetting_RoleMatrix(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)
	cnf.PAP.Store = "../../../testdata/apps/manager/policies/cedar"
	cnf.Authorization.FailClosedOnEmpty = true

	ip := newIDP(t)
	cnf.OIDC = config2.OIDC{JWKSURL: ip.srv.URL, Issuer: "https://idp.test", Audience: "openftv"}

	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))
	app := NewExternal(cnf, lgr).GetMainService().GetFiberApp()

	author, auditor, admin := ip.token(t, "author-user", "author"), ip.token(t, "auditor-user", "auditor"), ip.token(t, "admin-user", "admin")
	everyone := map[string]string{"author": author, "auditor": auditor, "admin": admin}

	tag := oas.Tag{Id: "laadpalen", Name: "Laadpalen"}
	ui := settings.Settings{HeaderTitle: "Vlierdam", HeaderColor: "#112233", TitleColor: "#AABBCC"}

	t.Run("everyone reads tags and settings", func(t *testing.T) {
		for name, tok := range everyone {
			require.Equal(t, http.StatusOK, call(t, app, http.MethodGet, "/v1/tags", tok, nil).StatusCode, name)
			require.Equal(t, http.StatusOK, call(t, app, http.MethodGet, "/v1/settings", tok, nil).StatusCode, name)
		}
	})

	t.Run("no token is denied, not served", func(t *testing.T) {
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodGet, "/v1/tags", "", nil).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodGet, "/v1/settings", "", nil).StatusCode)
	})

	t.Run("only the admin creates a tag", func(t *testing.T) {
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPost, "/v1/tag/laadpalen", author, tag).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPost, "/v1/tag/laadpalen", auditor, tag).StatusCode)
		require.Equal(t, http.StatusCreated, call(t, app, http.MethodPost, "/v1/tag/laadpalen", admin, tag).StatusCode)
	})

	t.Run("a denied caller cannot tell a missing tag from an existing one", func(t *testing.T) {
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodDelete, "/v1/tag/laadpalen", author, nil).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodDelete, "/v1/tag/nope", author, nil).StatusCode)
		require.Equal(t, http.StatusNotFound, call(t, app, http.MethodDelete, "/v1/tag/nope", admin, nil).StatusCode)
	})

	t.Run("only the admin updates and deletes a tag", func(t *testing.T) {
		tag.Name = "Laadpalen (bewerkt)"
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPut, "/v1/tag/laadpalen", author, tag).StatusCode)
		require.Equal(t, http.StatusOK, call(t, app, http.MethodPut, "/v1/tag/laadpalen", admin, tag).StatusCode)
		require.Equal(t, http.StatusOK, call(t, app, http.MethodDelete, "/v1/tag/laadpalen", admin, nil).StatusCode)
	})

	t.Run("only the admin writes the settings, without a path in any rule", func(t *testing.T) {
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPut, "/v1/settings", author, ui).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPut, "/v1/settings", auditor, ui).StatusCode)
		require.Equal(t, http.StatusOK, call(t, app, http.MethodPut, "/v1/settings", admin, ui).StatusCode)
	})
}
