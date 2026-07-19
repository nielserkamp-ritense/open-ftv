package pep

import (
	"io"
	"log/slog"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func TestWithJWTDefaultsRolesClaim(t *testing.T) {
	t.Parallel()

	p := New(nil, nil, WithJWT(JWTConfig{
		Keyfunc: func(*jwt.Token) (any, error) { return nil, nil },
		Issuer:  "https://idp.example/realms/openftv",
	}))

	if assert.NotNil(t, p.jwt) {
		assert.Equal(t, "roles", p.jwt.RolesClaim)
		assert.Equal(t, "https://idp.example/realms/openftv", p.jwt.Issuer)
	}
}

func TestExtractRoles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   any
		want []string
	}{
		{"nil", nil, nil},
		{"slice of any", []any{"admin", "auditor"}, []string{"admin", "auditor"}},
		{"slice of string", []string{"author"}, []string{"author"}},
		{"single string", "admin", []string{"admin"}},
		{"ignores non-strings", []any{"admin", 7, true}, []string{"admin"}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, extractRoles(tc.in))
		})
	}
}

func TestMapJWTPrincipal_PreferredUsername(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		claims jwt.MapClaims
		want   any
	}{
		{"claim present", jwt.MapClaims{"sub": "alice", "preferred_username": "Alice A."}, "Alice A."},
		{"claim absent", jwt.MapClaims{"sub": "alice"}, nil},
		{"claim empty", jwt.MapClaims{"sub": "alice", "preferred_username": ""}, nil},
		{"claim wrong type", jwt.MapClaims{"sub": "alice", "preferred_username": 123}, nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := &collector{
				logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
				jwt:    &JWTConfig{RolesClaim: "roles"},
				parc:   &models.PARC{Principal: models.NewEntity("", "", models.NewAttributeSet()), Context: models.NewAttributeSet()},
			}
			c.mapJWTPrincipal(tc.claims)

			assert.Equal(t, tc.want, c.parc.Principal.Attributes().GetAttributeValue(models.AttrPreferredName))
		})
	}
}
