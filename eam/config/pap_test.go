package config

import (
	"os"
	"testing"

	"gitlab.com/gjuyn/go-config/config-ext/yaml"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
)

func TestPAP(t *testing.T) {
	const badRecurse = `
policies:
  language: "cedar"
  store: 
    path: "/etc/cedar"
    recurse: "x"
`

	const goodCfg = `
policies:
  language: "REGO"
  store:
    path: "/etc/opa"
    recurse: true
`
	dir := t.TempDir()

	testCases := []struct {
		name         string
		file         string
		data         string
		args         []string
		wantErr      bool
		wantLanguage string
		wantStore    string
		wantRecurse  bool
	}{
		{
			name:    "bad input",
			file:    "pap1.yaml",
			data:    "%not yaml%",
			wantErr: true,
		},
		{
			name:         "bad recurse",
			file:         "pap2.yaml",
			data:         badRecurse,
			wantLanguage: "cedar",
			wantStore:    "/etc/cedar",
			wantRecurse:  false,
		},
		{
			name:         "good file",
			file:         "pap3.yaml",
			data:         goodCfg,
			wantLanguage: "REGO",
			wantStore:    "/etc/opa",
			wantRecurse:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file := dir + tc.file
			err := os.WriteFile(file, []byte(tc.data), 0644)

			cfg := &PAP{}

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
				assert.Equal(t, tc.wantLanguage, cfg.Language)
				assert.Equal(t, tc.wantStore, cfg.Store)
				assert.Equal(t, tc.wantRecurse, cfg.StoreRecurse)
			}
		})
	}
}
