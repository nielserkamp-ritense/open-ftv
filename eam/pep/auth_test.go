package pep

import (
	"crypto/rand"
	"crypto/rsa"
	"log/slog"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestProcessAuth(t *testing.T) {
	t.Parallel()

	token2 := jwt.New(jwt.SigningMethodPS256)

	testCases := []struct {
		name        string
		auth        string
		wantLog     int
		wantJWT     bool
		wantValid   bool
		wantHeaders map[string]any
		wantClaims  map[string]any
		wantUser    string
		wantPswd    string
	}{
		{
			name:    "empty",
			wantLog: 1,
		},
		{
			name:    "no type",
			auth:    "abcdef",
			wantLog: 1,
		},
		{
			name:    "invalid bearer token data",
			auth:    "Bearer abcdef",
			wantLog: 1,
		},
		{
			name:    "invalid bearer token signature",
			auth:    "Bearer " + signed2(token2),
			wantLog: 1,
		},
		{
			name:    "bad basic token encoding",
			auth:    "Basic ****",
			wantLog: 1,
		},
		{
			name:    "bad basic token format",
			auth:    "Basic YWRtaW4=",
			wantLog: 1,
		},
		{
			name:     "good basic token",
			auth:     "Basic bWlja2V5Om1vdXNl",
			wantUser: "mickey",
			wantPswd: "mouse",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := util.NewDummyHandler(slog.LevelDebug)

			c := &collector{
				debug:  true,
				logger: slog.New(h),
				req:    &models.HTTPRequest{},
				parc:   &models.PARC{Context: models.NewAttributeSet()},
			}
			c.processAuth(tc.auth)

			if tc.wantLog > 0 {
				assert.Equal(t, tc.wantLog, h.Count())
			}

			if tc.wantJWT {
				token, ok := c.parc.Context.GetAttributeValue(models.AttrJWT).(map[string]any)
				require.True(t, ok)
				require.NotNil(t, token)

				if tc.wantValid {
					got, ok2 := token[models.AttrValid].(bool)
					require.True(t, ok2)
					assert.True(t, got)
				}

				if tc.wantHeaders != nil {
					got, ok2 := token[models.AttrHeaders].(map[string]any)
					require.True(t, ok2)
					assert.EqualValues(t, tc.wantHeaders, got)
				}

				if tc.wantClaims != nil {
					got, ok2 := token[models.AttrClaims].(map[string]any)
					require.True(t, ok2)
					assert.EqualValues(t, tc.wantClaims, got)
				}
			}

			if tc.wantUser != "" {
				user, ok := c.parc.Context.GetAttributeValue(models.AttrBasicUser).(string)
				require.True(t, ok)
				assert.Equal(t, tc.wantUser, user)
			}

			if tc.wantPswd != "" {
				pswd, ok := c.parc.Context.GetAttributeValue(models.AttrBasicPswd).(string)
				require.True(t, ok)
				assert.Equal(t, tc.wantPswd, pswd)
			}
		})
	}
}

func signed2(token *jwt.Token) string {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	s, err2 := token.SignedString(key)
	if err2 != nil {
		panic(err2)
	}

	return s
}

// testKey is a fixed RSA key pair for signing and verifying test tokens.
func testKey(t *testing.T) (*rsa.PrivateKey, jwt.Keyfunc) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	kf := func(*jwt.Token) (any, error) { return &key.PublicKey, nil }
	return key, kf
}

func signRS256(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	s, err := tok.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func newCollector(kf jwt.Keyfunc) (*collector, *util.DummyHandler) {
	h := util.NewDummyHandler(slog.LevelDebug)
	c := &collector{
		debug:  true,
		logger: slog.New(h),
		req:    &models.HTTPRequest{},
		parc: &models.PARC{
			Principal: models.NewEntity("", "", models.NewAttributeSet()),
			Context:   models.NewAttributeSet(),
		},
		jwt: &JWTConfig{Keyfunc: kf, Issuer: "https://idp.test", Audience: "openftv", RolesClaim: "roles"},
	}
	return c, h
}

func TestProcessBearerValid(t *testing.T) {
	t.Parallel()
	key, kf := testKey(t)
	c, _ := newCollector(kf)

	bearer := signRS256(t, key, jwt.MapClaims{
		"iss":   "https://idp.test",
		"aud":   "openftv",
		"sub":   "alice",
		"roles": []any{"admin", "auditor"},
		"exp":   time.Now().Add(time.Hour).Unix(),
	})

	c.processAuth("Bearer " + bearer)
	c.determinePrincipal()

	assert.Equal(t, PrincipalUser, c.parc.Principal.Type())
	assert.Equal(t, "alice", c.parc.Principal.ID())
	assert.Equal(t, []string{"admin", "auditor"},
		c.parc.Principal.Attributes().GetAttributeValue(models.AttrRoles))
}

func TestProcessBearerRejectsBadSignature(t *testing.T) {
	t.Parallel()
	_, kf := testKey(t)
	other, _ := testKey(t)
	c, _ := newCollector(kf)

	bearer := signRS256(t, other, jwt.MapClaims{
		"iss": "https://idp.test", "aud": "openftv", "sub": "mallory",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	c.processAuth("Bearer " + bearer)
	c.determinePrincipal()

	assert.Equal(t, PrincipalInvalid, c.parc.Principal.ID())
}

func TestProcessBearerRejectsWrongIssuerAudienceExpiry(t *testing.T) {
	t.Parallel()
	key, kf := testKey(t)

	cases := map[string]jwt.MapClaims{
		"wrong issuer":   {"iss": "https://evil", "aud": "openftv", "sub": "x", "exp": time.Now().Add(time.Hour).Unix()},
		"wrong audience": {"iss": "https://idp.test", "aud": "other", "sub": "x", "exp": time.Now().Add(time.Hour).Unix()},
		"expired":        {"iss": "https://idp.test", "aud": "openftv", "sub": "x", "exp": time.Now().Add(-time.Hour).Unix()},
		"no expiry":      {"iss": "https://idp.test", "aud": "openftv", "sub": "x"},
	}
	for name, claims := range cases {
		claims := claims
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			c, _ := newCollector(kf)
			c.processAuth("Bearer " + signRS256(t, key, claims))
			c.determinePrincipal()
			assert.Equal(t, PrincipalInvalid, c.parc.Principal.ID())
		})
	}
}

func TestProcessBearerNotConfigured(t *testing.T) {
	t.Parallel()
	key, _ := testKey(t)
	c, h := newCollector(nil)
	c.jwt = nil

	c.processAuth("Bearer " + signRS256(t, key, jwt.MapClaims{
		"iss": "https://idp.test", "aud": "openftv", "sub": "x", "exp": time.Now().Add(time.Hour).Unix(),
	}))
	c.determinePrincipal()

	assert.Equal(t, PrincipalInvalid, c.parc.Principal.ID())
	assert.GreaterOrEqual(t, h.Count(), 1)
}
