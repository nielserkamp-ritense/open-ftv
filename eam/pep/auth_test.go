package pep

import (
	"crypto/rand"
	"crypto/rsa"
	"log/slog"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestProcessAuth(t *testing.T) {
	t.Parallel()

	token1 := jwt.New(jwt.SigningMethodHS256)
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
			name:        "good bearer token",
			auth:        "Bearer " + signed1(token1),
			wantJWT:     true,
			wantValid:   true,
			wantHeaders: map[string]any{"alg": "HS256", "typ": "JWT"},
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

func signed1(token *jwt.Token) string {
	key := []byte("secret1!")

	s, err := token.SignedString(key)
	if err != nil {
		panic(err)
	}

	return s
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
