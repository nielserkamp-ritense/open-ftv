package config

import (
	"os"
	"testing"

	"gitlab.com/gjuyn/go-config/config-ext/yaml"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
)

func TestOpenSearch(t *testing.T) {
	const goodCfg = `
opensearch:
  endpoints: "localhost:111,localhost:222,localhost:333"
  index: "authlog"
  user: "mickey"
  password: "mouse"
`
	dir := t.TempDir()

	testCases := []struct {
		name          string
		file          string
		data          string
		args          []string
		wantErr       bool
		wantEndpoints string
		wantIndex     string
		wantUser      string
		wantPswd      string
	}{
		{
			name:    "bad input",
			file:    "opensearch1.yaml",
			data:    "%not yaml%",
			wantErr: true,
		},
		{
			name:          "good file",
			file:          "opensearch2.yaml",
			data:          goodCfg,
			wantEndpoints: "localhost:111,localhost:222,localhost:333",
			wantIndex:     "authlog",
			wantUser:      "mickey",
			wantPswd:      "mouse",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file := dir + tc.file
			err := os.WriteFile(file, []byte(tc.data), 0644)

			cfg := &OpenSearch{}

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
				assert.Equal(t, tc.wantEndpoints, cfg.Endpoints)
				assert.Equal(t, tc.wantIndex, cfg.Index)
				assert.Equal(t, tc.wantUser, cfg.User)
				assert.Equal(t, tc.wantPswd, cfg.Pswd)
			}
		})
	}
}

func TestOpenSearch_Sanitized(t *testing.T) {
	t.Run("sanitize opensearch", func(t *testing.T) {
		o := &OpenSearch{
			Endpoints: "endpoints",
			Index:     "index",
			User:      "user",
			Pswd:      "password",
		}

		s := o.Sanitized()

		assert.Equal(t, "endpoints", s.Endpoints)
		assert.Equal(t, "index", s.Index)
		assert.Empty(t, s.User)
		assert.Empty(t, s.Pswd)
	})
}
