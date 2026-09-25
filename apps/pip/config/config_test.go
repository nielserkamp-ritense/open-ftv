package config

import (
	"testing"

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
			env:            map[string]string{"PIP_HOST": "127.0.0.1", "PIP_PORT": "8000", "PIP_LOG_LEVEL": "debug"},
			wantHost:       "127.0.0.1",
			wantPort:       9443,
			wantHealthHost: "127.0.0.1",
			wantLogLevel:   "warn",
		},
		{
			name:    "bad env value",
			env:     map[string]string{"PIP_READ_TIMEOUT": "soon"},
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

// TestLoad_DeployedEnv reads the environment variables a deployment of the PIP sets; none deploys it yet, so these mirror the Manager.
func TestLoad_DeployedEnv(t *testing.T) {
	t.Parallel()

	c, err := Load(nil, lookup(map[string]string{
		"PIP_HOST":                               "10.0.0.1",
		"PIP_PORT":                               "8444",
		"PIP_HEALTH_PORT":                        "8088",
		"PIP_LOG_LEVEL":                          "debug",
		"PIP_DATABASE_URL":                       "postgres://pip",
		"PIP_POLICIES_LANGUAGE":                  "cedar",
		"PIP_PIP_STORE":                          "/tmp/data/attributes",
		"PIP_PIP_PULL_CONFIGS":                   "/tmp/data/pull-config",
		"PIP_AUTHENTICATION_TYPE":                "bcrypt",
		"PIP_AUTHORIZATION_FAIL_CLOSED_ON_EMPTY": "true",
	}))
	require.NoError(t, err)

	assert.Equal(t, "10.0.0.1", c.Host)
	assert.Equal(t, "10.0.0.1", c.HealthHost)
	assert.Equal(t, uint16(8444), c.Port)
	assert.Equal(t, uint16(8088), c.HealthPort)
	assert.Equal(t, "debug", c.Level)
	assert.Equal(t, "postgres://pip", c.DatabaseURL)
	assert.Equal(t, "cedar", c.Language)
	assert.Equal(t, "/tmp/data/attributes", c.PIP.Store)
	assert.Equal(t, "/tmp/data/pull-config", c.PullConfigs)
	assert.Equal(t, "bcrypt", c.Type)
	assert.Equal(t, true, c.FailClosedOnEmpty)
}
