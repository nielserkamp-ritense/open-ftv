package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/gjuyn/go-config/config"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name       string
		args       []string
		wantErr    bool
		wantHost   string
		wantPort   uint16
		wantLogFmt string
		wantLogLvl string
		wantLogOut string
	}{
		{
			name:    "bad port",
			args:    []string{"--port=x"},
			wantErr: true,
		},
		{
			name:    "bad log file",
			args:    []string{"--log-output=/this/is/not/a/valid/path.log"},
			wantErr: true,
		},
		{
			name:       "good args",
			args:       []string{"--host=127.0.0.1", "--port=8080", "--log-format=text", "--log-level=warn", "--log-output=STDOUT", "--log-source=true"},
			wantHost:   "127.0.0.1",
			wantPort:   8080,
			wantLogFmt: "text",
			wantLogLvl: "warn",
			wantLogOut: "STDOUT",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, l, err2 := New(config.WithArgs(tc.args...))
			if tc.wantErr {
				require.Error(t, err2)
				require.Nil(t, c)
			} else {
				require.NoError(t, err2)
				require.NotNil(t, l)
				require.NotNil(t, c)

				assert.Equal(t, tc.wantHost, c.Host)
				assert.Equal(t, tc.wantPort, c.Port)
				assert.Equal(t, tc.wantLogFmt, c.Log.Format)
				assert.Equal(t, tc.wantLogLvl, c.Log.Level)
				assert.Equal(t, tc.wantLogOut, c.Log.Output)
				assert.True(t, c.Log.Source)
			}
		})
	}
}
