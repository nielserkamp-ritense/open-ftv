package config

import (
	"os"
	"testing"
	"time"

	"gitlab.com/gjuyn/go-config/config-ext/yaml"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
)

func TestOpenTel(t *testing.T) {
	const goodCfg = `
otel:
  url: "localhost:111"
  service: "authlog"
  timeout: "15s"
  prettyPrint: true
`
	dir := t.TempDir()

	testCases := []struct {
		name        string
		file        string
		data        string
		args        []string
		wantErr     bool
		wantURL     string
		wantService string
		wantTimeout time.Duration
		wantPretty  bool
	}{
		{
			name:    "bad input",
			file:    "opentel1.yaml",
			data:    "%not yaml%",
			wantErr: true,
		},
		{
			name:        "good file",
			file:        "opentel2.yaml",
			data:        goodCfg,
			wantURL:     "localhost:111",
			wantService: "authlog",
			wantTimeout: 15 * time.Second,
			wantPretty:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file := dir + tc.file
			err := os.WriteFile(file, []byte(tc.data), 0644)

			cfg := &OpenTel{}

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
				assert.Equal(t, tc.wantURL, cfg.URL)
				assert.Equal(t, tc.wantService, cfg.Service)
				assert.Equal(t, tc.wantTimeout, cfg.Timeout)
				assert.Equal(t, tc.wantPretty, cfg.Pretty)
			}
		})
	}
}
