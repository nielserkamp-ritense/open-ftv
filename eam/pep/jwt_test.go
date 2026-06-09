package pep

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
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
