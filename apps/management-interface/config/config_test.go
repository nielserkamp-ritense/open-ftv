package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		wantHealthPort uint16
	}{
		{
			name:           "defaults",
			wantHost:       "0.0.0.0",
			wantPort:       8080,
			wantHealthHost: "0.0.0.0",
			wantHealthPort: 8081,
		},
		{
			name:           "health host falls back to host",
			env:            map[string]string{"MANAGEMENT_INTERFACE_HOST": "127.0.0.1", "MANAGEMENT_INTERFACE_HEALTH_PORT": "9091"},
			args:           []string{"--port=9090"},
			wantHost:       "127.0.0.1",
			wantPort:       9090,
			wantHealthHost: "127.0.0.1",
			wantHealthPort: 9091,
		},
		{
			name:           "own health host",
			args:           []string{"--health-host=localhost"},
			wantHost:       "0.0.0.0",
			wantPort:       8080,
			wantHealthHost: "localhost",
			wantHealthPort: 8081,
		},
		{
			name:    "bad port",
			env:     map[string]string{"MANAGEMENT_INTERFACE_PORT": "x"},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := Load(tc.args, func(name string) (string, bool) {
				value, ok := tc.env[name]
				return value, ok
			})
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantHost, c.Host)
			assert.Equal(t, tc.wantPort, c.Port)
			assert.Equal(t, tc.wantHealthHost, c.HealthHost)
			assert.Equal(t, tc.wantHealthPort, c.HealthPort)
		})
	}
}
