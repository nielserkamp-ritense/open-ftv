//go:build integration

package server_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/server"
	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// idp is a throwaway OIDC issuer: a key pair and a JWKS endpoint the manager trusts.
type idp struct {
	key *rsa.PrivateKey
	srv *httptest.Server
}

func newIDP(t *testing.T) *idp {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	jwks, err := json.Marshal(map[string]any{"keys": []map[string]any{{
		"kty": "RSA", "kid": "test", "use": "sig", "alg": "RS256",
		"n": base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
		"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes()),
	}}})
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwks)
	}))
	t.Cleanup(srv.Close)

	return &idp{key: key, srv: srv}
}

// token mints an access token for sub with the given roles, as the shipped realm would.
func (i *idp) token(t *testing.T, sub string, roles ...string) string {
	t.Helper()

	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":                "https://idp.test",
		"aud":                "openftv",
		"sub":                sub,
		"preferred_username": sub,
		"roles":              roles,
		"exp":                time.Now().Add(time.Hour).Unix(),
	})
	tok.Header["kid"] = "test"

	bearer, err := tok.SignedString(i.key)
	require.NoError(t, err)

	return bearer
}

func call(t *testing.T, app *fiber.App, method, path, bearer string, body any) *http.Response {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")

	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}

	res, err := app.Test(req, 15000)
	require.NoError(t, err)

	return res
}

// Test_Policy_RoleMatrix drives the SCRUM-16 matrix for policies through the manager as
// deployed: secured mode, OIDC, the shipped policies seeded into postgres.
func Test_Policy_RoleMatrix(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)
	cnf.PAP.Store = "../../../testdata/apps/manager/policies/cedar"
	cnf.Authorization.FailClosedOnEmpty = true

	ip := newIDP(t)
	cnf.OIDC = config2.OIDC{JWKSURL: ip.srv.URL, Issuer: "https://idp.test", Audience: "openftv"}

	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))
	app := server.NewExternal(cnf, lgr).GetMainService().GetFiberApp()

	author, auditor, admin := ip.token(t, "author-user", "author"), ip.token(t, "auditor-user", "auditor"), ip.token(t, "admin-user", "admin")

	const id = "44444444-4444-4444-4444-444444444444"
	// Every stored policy is enforced by the manager's own PDP as well, so the policy under
	// administration must be one that matches nothing on the management plane.
	policy := oas.Policy{Language: "cedar", Data: "permit (principal is doelbinding, action, resource is service);", Metadata: oas.Metadata{Title: "Matrix"}}

	t.Run("everyone reads the collection", func(t *testing.T) {
		for name, tok := range map[string]string{"author": author, "auditor": auditor, "admin": admin} {
			require.Equal(t, http.StatusOK, call(t, app, http.MethodGet, "/v1/policies", tok, nil).StatusCode, name)
		}
	})

	t.Run("no token is denied, not served", func(t *testing.T) {
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodGet, "/v1/policies", "", nil).StatusCode)
	})

	t.Run("only the author creates", func(t *testing.T) {
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPost, "/v1/policy/"+id, auditor, policy).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPost, "/v1/policy/"+id, admin, policy).StatusCode)
		require.Equal(t, http.StatusCreated, call(t, app, http.MethodPost, "/v1/policy/"+id, author, policy).StatusCode)
	})

	t.Run("a denied caller cannot tell a missing id from an existing one", func(t *testing.T) {
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodDelete, "/v1/policy/"+id, auditor, nil).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodDelete, "/v1/policy/55555555-5555-5555-5555-555555555555", auditor, nil).StatusCode)
	})

	t.Run("status changes are accept and deploy, never a read", func(t *testing.T) {
		accepted := oas.PolicyStatus{Id: id, Status: "accepted"}

		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPatch, "/v1/policy/"+id+"/status", auditor, accepted).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPatch, "/v1/policy/"+id+"/status", admin, accepted).StatusCode)
		require.Equal(t, http.StatusOK, call(t, app, http.MethodPatch, "/v1/policy/"+id+"/status", author, accepted).StatusCode)
	})

	t.Run("an accepted policy is not edited in place", func(t *testing.T) {
		policy.Metadata.Title = "Edited"
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPut, "/v1/policy/"+id, author, policy).StatusCode)
	})

	t.Run("only the author reverts, and an illegal transition is a client error", func(t *testing.T) {
		concept := oas.PolicyStatus{Id: id, Status: "concept"}

		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPatch, "/v1/policy/"+id+"/status", admin, concept).StatusCode)
		require.Equal(t, http.StatusBadRequest, call(t, app, http.MethodPatch, "/v1/policy/"+id+"/status", author, oas.PolicyStatus{Id: id, Status: "deployed"}).StatusCode)
		require.Equal(t, http.StatusOK, call(t, app, http.MethodPatch, "/v1/policy/"+id+"/status", author, concept).StatusCode)

		// back in concept, the author edits again.
		require.Equal(t, http.StatusOK, call(t, app, http.MethodPut, "/v1/policy/"+id, author, policy).StatusCode)
	})

	t.Run("the author is attributed", func(t *testing.T) {
		res := call(t, app, http.MethodGet, "/v1/policy/"+id, auditor, nil)
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Equal(t, "author-user", decodePolicy(t, res).Audit.CreatedBy.Id)
	})
}
