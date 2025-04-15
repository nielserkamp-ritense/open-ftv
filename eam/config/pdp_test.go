package config

import (
	"os"
	"testing"

	"gitlab.com/gjuyn/go-config/config-ext/yaml"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
)

func TestPDP(t *testing.T) {
	t.Parallel()

	const goodCfg = `
pdp:
  request:
    mappings: "a,b,c"
`
	dir := t.TempDir()

	testCases := []struct {
		name        string
		file        string
		data        string
		args        []string
		wantErr     bool
		wantMapping string
	}{
		{
			name:    "bad input",
			file:    "pdp1.yaml",
			data:    "%not yaml%",
			wantErr: true,
		},
		{
			name:        "good file",
			file:        "pdp2.yaml",
			data:        goodCfg,
			wantMapping: "a,b,c",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := dir + tc.file
			err := os.WriteFile(file, []byte(tc.data), 0644)

			cfg := &PDP{}

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
				assert.Equal(t, tc.wantMapping, cfg.RequestMappings)
			}
		})
	}
}
