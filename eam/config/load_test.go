package config

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
	"gitlab.com/gjuyn/go-config/config-ext/yaml"
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

	var goodCfg = `
svc:
  host: "127.0.0.1"
  port: 8080
  maxBody: 2048
log:
  format: "text"
  level: "info"
  output: "STDOUT"
  source: true
`

	var badLogger = `
svc:
  host: "127.0.0.1"
  port: 8080
  maxBody: 2048
log:
  format: "json"
  level: "debug"
  output: "/this/is/not/a/valid/path.log"
`
	dir := t.TempDir()

	testCases := []struct {
		name        string
		file        string
		data        string
		wantErr     bool
		wantHost    string
		wantPort    uint16
		wantMaxBody int
		wantLogFmt  string
		wantLogLvl  string
		wantLogOut  string
	}{
		{
			name:    "bad input",
			file:    "load1.yaml",
			data:    "%not yaml%",
			wantErr: true,
		},
		{
			name:    "bad log file",
			file:    "load2.yaml",
			data:    badLogger,
			wantErr: true,
		},
		{
			name:        "good file",
			file:        "load3.yaml",
			data:        goodCfg,
			wantHost:    "127.0.0.1",
			wantPort:    8080,
			wantMaxBody: 2048,
			wantLogFmt:  "text",
			wantLogLvl:  "info",
			wantLogOut:  "STDOUT",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := dir + tc.file
			err := os.WriteFile(file, []byte(tc.data), 0644)
			require.NoError(t, err)

			defer func() {
				e := recover()
				if tc.wantErr {
					require.NotNil(t, e)
				} else {
					require.Nil(t, e)
				}
			}()

			cfg := TestConfig{}

			l := Load(&cfg, config.NoFlags(), yaml.FilesYAML(file))
			if tc.wantErr {
				require.True(t, false) // cannot happen
			}

			require.NotNil(t, l)

			assert.Equal(t, tc.wantHost, cfg.Server.Host)
			assert.Equal(t, tc.wantPort, cfg.Server.Port)
			assert.Equal(t, tc.wantMaxBody, cfg.Server.MaxBody)
			assert.Equal(t, tc.wantLogFmt, cfg.Log.Format)
			assert.Equal(t, tc.wantLogLvl, cfg.Log.Level)
			assert.Equal(t, tc.wantLogOut, cfg.Log.Output)
			assert.True(t, cfg.Log.Source)
		})
	}
}
