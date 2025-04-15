package config

import (
	"os"
	"testing"

	"gitlab.com/gjuyn/go-config/config-ext/yaml"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
)

func TestServer(t *testing.T) {
	t.Parallel()

	const badPort = `
svc:
  host: "127.0.0.1"
  port: "x"
  maxBody: 1111
`

	const goodCfg = `
svc:
  host: "127.0.0.1"
  port: 8080
  maxBody: 2048
`
	dir := t.TempDir()

	testCases := []struct {
		name        string
		file        string
		data        string
		args        []string
		wantErr     bool
		wantHost    string
		wantPort    uint16
		wantMaxBody int
	}{
		{
			name:    "bad input",
			file:    "server1.yaml",
			data:    "%not yaml%",
			wantErr: true,
		},
		{
			name:        "bad port",
			file:        "server2.yaml",
			data:        badPort,
			wantHost:    "127.0.0.1",
			wantPort:    0,
			wantMaxBody: 1111,
		},
		{
			name:        "good file",
			file:        "server3.yaml",
			data:        goodCfg,
			wantHost:    "127.0.0.1",
			wantPort:    8080,
			wantMaxBody: 2048,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := dir + tc.file
			err := os.WriteFile(file, []byte(tc.data), 0644)

			cfg := &Server{}

			opts := []config.Option{
				yaml.FilesYAML(file),
				config.WithArgs(tc.args...),
				config.EnvironmentPrefix("CFG_"),
				config.AppName("test 1.0"),
				config.NoHelpOnError(),
			}

			err = config.LoadConfig(cfg, opts...)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantHost, cfg.Host)
				assert.Equal(t, tc.wantPort, cfg.Port)
				assert.Equal(t, tc.wantMaxBody, cfg.MaxBody)
			}
		})
	}
}
