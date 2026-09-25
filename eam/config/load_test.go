package config

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
)

type TestConfig struct {
	Server
	Log
}

func (c *TestConfig) LogSanitized(logger *slog.Logger) {
	logger.Info("config", "values", c)
}

func TestLoad(t *testing.T) {
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

			cfg := TestConfig{}

			l := Load(&cfg, &cfg.Log, config.WithArgs(tc.args...))
			if tc.wantErr {
				require.True(t, false) // cannot happen
			}

			require.NotNil(t, l)

			assert.Equal(t, tc.wantHost, cfg.Server.Host)
			assert.Equal(t, tc.wantPort, cfg.Server.Port)
			assert.Equal(t, tc.wantLogFmt, cfg.Log.Format)
			assert.Equal(t, tc.wantLogLvl, cfg.Log.Level)
			assert.Equal(t, tc.wantLogOut, cfg.Log.Output)
			assert.True(t, cfg.Log.Source)
		})
	}
}
