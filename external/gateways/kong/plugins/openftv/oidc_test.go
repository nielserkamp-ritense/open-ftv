package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func mustParse(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	require.NoError(t, err)
	return u
}

func TestPEPOptionFromConfigValidatesJWKSSignedToken(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	const kid = "test-key-1"
	jwks := map[string]any{"keys": []map[string]any{{
		"kty": "RSA", "alg": "RS256", "use": "sig", "kid": kid,
		"n": base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
		"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes()),
	}}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer srv.Close()

	cfg := &Config{Issuer: "https://idp.test", JWKSURL: srv.URL, Audience: "openftv", RolesClaim: "roles"}
	opt, err := pepOptionFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, opt)

	p := newPEPWith(opt)

	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": "https://idp.test", "aud": "openftv", "sub": "carol",
		"roles": []any{"admin"}, "exp": time.Now().Add(time.Hour).Unix(),
	})
	tok.Header["kid"] = kid
	signed, err := tok.SignedString(key)
	require.NoError(t, err)

	now := time.Now().UTC()
	req := models.Request{
		RequestTime: &now,
		Method:      "GET",
		URL:         mustParse(t, "https://mgr.test/v1/policies"),
		Headers:     map[string][]string{models.HeaderAuthorization: {"Bearer " + signed}},
	}
	parc := p.PARCFromRequest(&req, func(string) (*models.Entity, uint64, error) { return nil, 0, nil })

	assert.Equal(t, "user::carol", parc.Principal.UID())
	assert.Equal(t, []string{"admin"}, parc.Principal.Attributes().GetAttributeValue(models.AttrRoles))
}

func TestEnsurePEPValidatesToken(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	const kid = "test-key-1"
	jwks := map[string]any{"keys": []map[string]any{{
		"kty": "RSA", "alg": "RS256", "use": "sig", "kid": kid,
		"n": base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
		"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes()),
	}}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer srv.Close()

	c := &Config{Issuer: "https://idp.test", JWKSURL: srv.URL, Audience: "openftv", RolesClaim: "roles"}
	c.ensurePEP()
	c.ensurePEP() // idempotent; must not panic or rebuild

	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": "https://idp.test", "aud": "openftv", "sub": "carol",
		"roles": []any{"admin"}, "exp": time.Now().Add(time.Hour).Unix(),
	})
	tok.Header["kid"] = kid
	signed, err := tok.SignedString(key)
	require.NoError(t, err)

	now := time.Now().UTC()
	req := models.Request{
		RequestTime: &now, Method: "GET", URL: mustParse(t, "https://mgr.test/v1/policies"),
		Headers: map[string][]string{models.HeaderAuthorization: {"Bearer " + signed}},
	}
	parc := c.pep.PARCFromRequest(&req, func(string) (*models.Entity, uint64, error) { return nil, 0, nil })
	assert.Equal(t, "user::carol", parc.Principal.UID())
	assert.Equal(t, []string{"admin"}, parc.Principal.Attributes().GetAttributeValue(models.AttrRoles))
}
