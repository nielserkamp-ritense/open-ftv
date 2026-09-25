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
			env:            map[string]string{"PAP_HOST": "127.0.0.1", "PAP_PORT": "8000", "PAP_LOG_LEVEL": "debug"},
			wantHost:       "127.0.0.1",
			wantPort:       9443,
			wantHealthHost: "127.0.0.1",
			wantLogLevel:   "warn",
		},
		{
			name:    "bad env value",
			env:     map[string]string{"PAP_READ_TIMEOUT": "soon"},
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

// TestLoad_DeployedEnv reads the environment variables a deployment of the PAP sets; none deploys it yet, so these mirror the Manager.
func TestLoad_DeployedEnv(t *testing.T) {
	t.Parallel()

	c, err := Load(nil, lookup(map[string]string{
		"PAP_HOST":                       "10.0.0.1",
		"PAP_PORT":                       "8444",
		"PAP_INTERNAL_PORT":              "9000",
		"PAP_LOG_LEVEL":                  "debug",
		"PAP_DATABASE_URL":               "postgres://pap",
		"PAP_POLICIES_LANGUAGE":          "cedar",
		"PAP_POLICIES_STORE":             "/tmp/data/policies",
		"PAP_TAGS_PATH":                  "/tmp/data/tags",
		"PAP_AUTHENTICATION_TYPE":        "bcrypt",
		"PAP_BUNDLE_CONFIGS":             "/tmp/data/bundles",
		"PAP_BUNDLE_SEND_TIMEOUT":        "2m",
		"PAP_AUTHORIZATION_AUTHENTICATE": "true",
	}))
	require.NoError(t, err)

	assert.Equal(t, "10.0.0.1", c.Host)
	assert.Equal(t, "10.0.0.1", c.HealthHost)
	assert.Equal(t, uint16(8444), c.Port)
	assert.Equal(t, uint16(9000), c.InternalPort)
	assert.Equal(t, "debug", c.Level)
	assert.Equal(t, "postgres://pap", c.DatabaseURL)
	assert.Equal(t, "cedar", c.Language)
	assert.Equal(t, "/tmp/data/policies", c.PAP.Store)
	assert.Equal(t, "/tmp/data/tags", c.TagsPath)
	assert.Equal(t, "bcrypt", c.Type)
	assert.Equal(t, "/tmp/data/bundles", c.BundlePath)
	assert.Equal(t, 2*time.Minute, c.BundleTimeout)
	assert.Equal(t, true, c.Authenticate)
}
