package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		args    []string
		env     map[string]string
		wantErr bool
		want    *cfg
	}{
		{
			name: "defaults",
			want: &cfg{Host: "0.0.0.0", Port: 3001, PDP: "https://localhost:8443/authzen/v1/evaluation", Timeout: 5 * time.Second},
		},
		{
			name: "env and flags",
			args: []string{"--port=9000"},
			env:  map[string]string{"ENVOY_AUTHZEN_PDP_ADDRESS": "http://envoy-pdp:8443/authzen/v1/evaluation", "ENVOY_AUTHZEN_PORT": "8000"},
			want: &cfg{Host: "0.0.0.0", Port: 9000, PDP: "http://envoy-pdp:8443/authzen/v1/evaluation", Timeout: 5 * time.Second},
		},
		{
			name:    "bad timeout",
			env:     map[string]string{"ENVOY_AUTHZEN_PDP_TIMEOUT": "soon"},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := loadConfig(tc.args, func(name string) (string, bool) {
				value, ok := tc.env[name]
				return value, ok
			})
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
