package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrincipal_IsUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    Principal
		want bool
	}{
		{"user", NewPrincipal(KindUser, "alice"), true},
		{"user with empty id", NewPrincipal(KindUser, ""), false},
		{"app", NewPrincipal(KindApp, "abcdef"), false},
		{"system", NewSystemPrincipal(), false},
		{"invalid", NewPrincipal(KindInvalid, "invalid"), false},
		{"zero value", Principal{}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.p.IsAuthenticatedUser())
		})
	}
}
