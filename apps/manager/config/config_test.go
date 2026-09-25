package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
)

func TestNew(t *testing.T) {
	t.Parallel()

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
			args:       []string{"--host=127.0.0.1", "--port=8080", "--log-format=text", "--log-level=info", "--log-output=STDOUT", "--log-source=true"},
			wantHost:   "127.0.0.1",
			wantPort:   8080,
			wantLogFmt: "text",
			wantLogLvl: "info",
			wantLogOut: "STDOUT",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				e := recover()
				if tc.wantErr {
					require.NotNil(t, e)
				} else {
					require.Nil(t, e)
				}
			}()

			c, l := New(config.WithArgs(tc.args...))
			if tc.wantErr {
				require.Nil(t, c)
				require.Nil(t, l)
			} else {
				require.NotNil(t, c)
				require.NotNil(t, l)

				assert.Equal(t, tc.wantHost, c.Host)
				assert.Equal(t, tc.wantPort, c.Port)
				assert.Equal(t, tc.wantLogFmt, c.Format)
				assert.Equal(t, tc.wantLogLvl, c.Level)
				assert.Equal(t, tc.wantLogOut, c.Output)
				assert.True(t, c.Log.Source)
			}
		})
	}
}
