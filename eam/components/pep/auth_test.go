package pep

import (
	"log/slog"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestProcessAuth(t *testing.T) {
	testCases := []struct {
		name        string
		auth        string
		wantLog     int
		wantJWT     bool
		wantValid   bool
		wantHeaders map[string]any
		wantClaims  map[string]any
	}{
		{
			name:    "empty",
			wantLog: 1,
		},
		{
			name:    "no bearer",
			auth:    "abcdef",
			wantLog: 1,
		},
		{
			name:    "invalid token",
			auth:    "Bearer abcdef",
			wantLog: 1,
		},
		{
			name:        "basic token",
			auth:        "Bearer " + signed(token1),
			wantJWT:     true,
			wantValid:   true,
			wantHeaders: map[string]any{"alg": "HS256", "typ": "JWT"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
		})
	}
}

func signed(token *jwt.Token) string {
	s, err := token.SignedString(key1)
	if err != nil {
		panic(err)
	}
	return s
}

var (
	key1   = []byte("secret1!")
	token1 = jwt.New(jwt.SigningMethodHS256)
)
