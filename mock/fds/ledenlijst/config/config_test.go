package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		args       []string
		env        map[string]string
		wantErr    bool
		wantHost   string
		wantPort   uint16
		wantLogFmt string
		wantLogLvl string
	}{
		{
			name:       "defaults",
			wantHost:   "0.0.0.0",
			wantPort:   8090,
			wantLogFmt: "json",
			wantLogLvl: "info",
		},
		{
			name:       "env and flags",
			args:       []string{"--host=127.0.0.1", "--log-level=warn"},
			env:        map[string]string{"FDS_LEDENLIJST_PORT": "8080", "FDS_LEDENLIJST_LOG_FORMAT": "text", "FDS_LEDENLIJST_LOG_LEVEL": "debug"},
			wantHost:   "127.0.0.1",
			wantPort:   8080,
			wantLogFmt: "text",
			wantLogLvl: "warn",
		},
		{
			name:    "bad port",
			args:    []string{"--port=x"},
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
			assert.Equal(t, tc.wantLogFmt, c.Log.Format)
			assert.Equal(t, tc.wantLogLvl, c.Log.Level)
			assert.True(t, c.Log.Source)
		})
	}
}
