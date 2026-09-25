package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lookup(env map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		value, ok := env[name]
		return value, ok
	}
}

func TestLoad(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		args           []string
		env            map[string]string
		wantErr        bool
		wantHost       string
		wantPort       uint16
		wantHealthHost string
		wantLogLevel   string
	}{
		{
			name:           "defaults",
			wantHost:       "0.0.0.0",
			wantPort:       8443,
			wantHealthHost: "0.0.0.0",
			wantLogLevel:   "info",
		},
		{
			name:           "flag overrides env",
			args:           []string{"--port=9443", "--log-level=warn"},
			env:            map[string]string{"MANAGER_HOST": "127.0.0.1", "MANAGER_PORT": "8000", "MANAGER_LOG_LEVEL": "debug"},
			wantHost:       "127.0.0.1",
			wantPort:       9443,
			wantHealthHost: "127.0.0.1",
			wantLogLevel:   "warn",
		},
		{
			name:    "bad env value",
			env:     map[string]string{"MANAGER_READ_TIMEOUT": "soon"},
			wantErr: true,
		},
		{
			name:    "unknown flag",
			args:    []string{"--address=127.0.0.1"},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := Load(tc.args, lookup(tc.env))
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantHost, c.Host)
			assert.Equal(t, tc.wantPort, c.Port)
			assert.Equal(t, tc.wantHealthHost, c.HealthHost)
			assert.Equal(t, tc.wantLogLevel, c.Level)
		})
	}
}

// TestLoad_DeployedEnv reads the environment variables the Helm chart and compose files set for the Manager.
func TestLoad_DeployedEnv(t *testing.T) {
	t.Parallel()

	c, err := Load(nil, lookup(map[string]string{
		"MANAGER_HOST":                               "10.0.0.1",
		"MANAGER_PORT":                               "8444",
		"MANAGER_CORS_ORIGINS":                       "https://ui",
		"MANAGER_CORS_HEADERS":                       "Authorization",
		"MANAGER_HEALTH_PORT":                        "8088",
		"MANAGER_LOG_FORMAT":                         "text",
		"MANAGER_LOG_LEVEL":                          "debug",
		"MANAGER_DATABASE_URL":                       "postgres://manager",
		"MANAGER_ADL_PG_URL":                         "postgres://adl",
		"MANAGER_OIDC_ISSUER":                        "https://idp/realms/ftv",
		"MANAGER_OIDC_JWKS_URL":                      "https://idp/certs",
		"MANAGER_OIDC_AUDIENCE":                      "manager",
		"MANAGER_OIDC_ROLES_CLAIM":                   "roles",
		"MANAGER_POLICIES_LANGUAGE":                  "cedar",
		"MANAGER_POLICIES_STORE":                     "/tmp/data/policies",
		"MANAGER_POLICIES_STORE_RECURSE":             "true",
		"MANAGER_PIP_STORE":                          "/tmp/data/attributes",
		"MANAGER_PIP_STORE_RECURSE":                  "true",
		"MANAGER_TAGS_PATH":                          "/tmp/data/tags",
		"MANAGER_AUTHENTICATION_TYPE":                "bcrypt",
		"MANAGER_BUNDLE_CONFIGS":                     "/tmp/data/bundles",
		"MANAGER_BUNDLE_RECURSE":                     "true",
		"MANAGER_BUNDLE_SEND_TIMEOUT":                "2m",
		"MANAGER_AUTHORIZATION_FAIL_CLOSED_ON_EMPTY": "true",
	}))
	require.NoError(t, err)

	assert.Equal(t, "10.0.0.1", c.Host)
	assert.Equal(t, "10.0.0.1", c.HealthHost)
	assert.Equal(t, uint16(8444), c.Port)
	assert.Equal(t, "https://ui", c.CorsOrigins)
	assert.Equal(t, "Authorization", c.CorsHeaders)
	assert.Equal(t, uint16(8088), c.HealthPort)
	assert.Equal(t, "text", c.Format)
	assert.Equal(t, "debug", c.Level)
	assert.Equal(t, "postgres://manager", c.DatabaseURL)
	assert.Equal(t, "postgres://adl", c.ADLDatabaseURL)
	assert.Equal(t, "https://idp/realms/ftv", c.Issuer)
	assert.Equal(t, "https://idp/certs", c.JWKSURL)
	assert.Equal(t, "manager", c.Audience)
	assert.Equal(t, "roles", c.RolesClaim)
	assert.Equal(t, "cedar", c.Language)
	assert.Equal(t, "/tmp/data/policies", c.PAP.Store)
	assert.Equal(t, true, c.PAP.StoreRecurse)
	assert.Equal(t, "/tmp/data/attributes", c.PIP.Store)
	assert.Equal(t, true, c.PIP.StoreRecurse)
	assert.Equal(t, "/tmp/data/tags", c.TagsPath)
	assert.Equal(t, "bcrypt", c.Type)
	assert.Equal(t, "/tmp/data/bundles", c.BundlePath)
	assert.Equal(t, true, c.BundleRecurse)
	assert.Equal(t, 2*time.Minute, c.BundleTimeout)
	assert.Equal(t, true, c.FailClosedOnEmpty)
}
