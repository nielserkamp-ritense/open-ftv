package config

import (
	"os"
	"testing"

	"gitlab.com/gjuyn/go-config/config-ext/yaml"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
)

func TestLog(t *testing.T) {
	t.Parallel()

	const badSource = `
log:
  output: "STDOUT"
  format: "json"
  level: "debug"
  source: "x"
`

	const goodCfg = `
log:
  output: "/logs/mylog.txt"
  format: "TEXT"
  level: "info"
  source: true
`
	dir := t.TempDir()

	testCases := []struct {
		name       string
		file       string
		data       string
		args       []string
		wantErr    bool
		wantOutput string
		wantFormat string
		wantLevel  string
		wantSource bool
	}{
		{
			name:    "bad input",
			file:    "log1.yaml",
			data:    "%not yaml%",
			wantErr: true,
		},
		{
			name:       "bad source",
			file:       "log2.yaml",
			data:       badSource,
			wantOutput: "STDOUT",
			wantFormat: "json",
			wantLevel:  "debug",
			wantSource: false,
		},
		{
			name:       "good file",
			file:       "log3.yaml",
			data:       goodCfg,
			wantOutput: "/logs/mylog.txt",
			wantFormat: "TEXT",
			wantLevel:  "info",
			wantSource: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := dir + tc.file
			err := os.WriteFile(file, []byte(tc.data), 0644)

			cfg := &Log{}

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
				assert.Equal(t, tc.wantOutput, cfg.Output)
				assert.Equal(t, tc.wantFormat, cfg.Format)
				assert.Equal(t, tc.wantLevel, cfg.Level)
				assert.Equal(t, tc.wantSource, cfg.Source)
			}
		})
	}
}
