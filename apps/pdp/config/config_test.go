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
			env:            map[string]string{"PDP_HOST": "127.0.0.1", "PDP_PORT": "8000", "PDP_LOG_LEVEL": "debug"},
			wantHost:       "127.0.0.1",
			wantPort:       9443,
			wantHealthHost: "127.0.0.1",
			wantLogLevel:   "warn",
		},
		{
			name:    "bad env value",
			env:     map[string]string{"PDP_READ_TIMEOUT": "soon"},
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

// TestLoad_DeployedEnv reads the environment variables the Helm chart and compose files set for the PDP.
func TestLoad_DeployedEnv(t *testing.T) {
	t.Parallel()

	c, err := Load(nil, lookup(map[string]string{
		"PDP_HOST":                   "10.0.0.1",
		"PDP_PORT":                   "8444",
		"PDP_TLS_CERT":               "/var/run/secrets/open-ftv/pdp/tls.crt",
		"PDP_TLS_KEY":                "/var/run/secrets/open-ftv/pdp/tls.key",
		"PDP_HEALTH_PORT":            "8088",
		"PDP_LOG_FORMAT":             "text",
		"PDP_LOG_LEVEL":              "debug",
		"PDP_POLICIES_LANGUAGE":      "cedar",
		"PDP_POLICIES_STORE":         "/etc/open-ftv/pdp/policies",
		"PDP_POLICIES_STORE_RECURSE": "true",
		"PDP_PIP_STORE":              "/etc/open-ftv/pdp/attributes",
		"PDP_PIP_PULL_CONFIGS":       "/etc/open-ftv/pdp/pip-pull-config.yaml",
		"PDP_REQUEST_MAPPINGS":       "a,b",
		"PDP_BUNDLE_MANAGER":         "http://manager:9443/bundles",
		"PDP_BUNDLE_ENCODING":        "bzip2",
		"PDP_BUNDLE_TIMEOUT":         "2m",
		"PDP_BUNDLE_RETRIES":         "3",
		"PDP_ADL_SERVICE_NAME":       "pdp1",
		"PDP_ADL_PG_URL":             "postgres://adl",
	}))
	require.NoError(t, err)

	assert.Equal(t, "10.0.0.1", c.Host)
	assert.Equal(t, "10.0.0.1", c.HealthHost)
	assert.Equal(t, uint16(8444), c.Port)
	assert.Equal(t, "/var/run/secrets/open-ftv/pdp/tls.crt", c.Cert)
	assert.Equal(t, "/var/run/secrets/open-ftv/pdp/tls.key", c.Key)
	assert.Equal(t, uint16(8088), c.HealthPort)
	assert.Equal(t, "text", c.Format)
	assert.Equal(t, "debug", c.Level)
	assert.Equal(t, "cedar", c.Language)
	assert.Equal(t, "/etc/open-ftv/pdp/policies", c.PAP.Store)
	assert.Equal(t, true, c.PAP.StoreRecurse)
	assert.Equal(t, "/etc/open-ftv/pdp/attributes", c.PIP.Store)
	assert.Equal(t, "/etc/open-ftv/pdp/pip-pull-config.yaml", c.PullConfigs)
	assert.Equal(t, "a,b", c.RequestMappings)
	assert.Equal(t, "http://manager:9443/bundles", c.BundleManager)
	assert.Equal(t, "bzip2", c.BundleEncoding)
	assert.Equal(t, 2*time.Minute, c.BundleTimeout)
	assert.Equal(t, 3, c.BundleRetries)
	assert.Equal(t, "pdp1", c.ADLService)
	assert.Equal(t, "postgres://adl", c.ADLDatabaseURL)
}
