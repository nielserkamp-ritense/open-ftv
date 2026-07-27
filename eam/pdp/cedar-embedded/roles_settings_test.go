package cedar_embedded

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSettingsAuthorization verifies settings_admin_only.cedar.
func TestSettingsAuthorization(t *testing.T) {
	c := managerPDP(t) // loads testdata/apps/manager/policies/cedar (incl. settings_admin_only.cedar)

	cases := []struct {
		role, action, path string
		want               bool
	}{
		{"admin", "can_create", "/v1/settings", true},
		{"author", "can_create", "/v1/settings", false},
		{"auditor", "can_create", "/v1/settings", false},
		{"admin", "can_read", "/v1/settings", true},
		{"author", "can_read", "/v1/settings", true},
		{"auditor", "can_read", "/v1/settings", true},
	}
	for _, tc := range cases {
		t.Run(tc.role+"-"+tc.action, func(t *testing.T) {
			resp, err := c.Authorize("uid", parcFor(tc.role, tc.action, tc.path))
			require.NoError(t, err)
			assert.Equal(t, tc.want, resp.Allowed, "%s %s %s", tc.role, tc.action, tc.path)
		})
	}
}
