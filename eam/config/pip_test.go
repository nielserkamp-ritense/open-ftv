package config

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"gitlab.com/gjuyn/go-config/config-ext/yaml"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestPIP(t *testing.T) {
	t.Parallel()

	const badRecurse = `
pip:
  store: 
    path: "/etc/pip"
    recurse: "x"
  pull:
    configPath: "/etc/pip/pull"
`

	const goodCfg = `
pip:
  store:
    path: "/etc/pip"
    recurse: true
  pull:
    configPath: "/etc/pip/pull"
`
	dir := t.TempDir()

	testCases := []struct {
		name        string
		file        string
		data        string
		args        []string
		wantErr     bool
		wantStore   string
		wantRecurse bool
		wantPullCfg string
	}{
		{
			name:    "bad input",
			file:    "pap1.yaml",
			data:    "%not yaml%",
			wantErr: true,
		},
		{
			name:        "bad recurse",
			file:        "pap2.yaml",
			data:        badRecurse,
			wantStore:   "/etc/pip",
			wantRecurse: false,
			wantPullCfg: "/etc/pip/pull",
		},
		{
			name:        "good file",
			file:        "pap3.yaml",
			data:        goodCfg,
			wantStore:   "/etc/pip",
			wantRecurse: true,
			wantPullCfg: "/etc/pip/pull",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := dir + tc.file
			err := os.WriteFile(file, []byte(tc.data), 0644)

			cfg := &PIP{}

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
				assert.Equal(t, tc.wantStore, cfg.Store)
				assert.Equal(t, tc.wantRecurse, cfg.StoreRecurse)
				assert.Equal(t, tc.wantPullCfg, cfg.PullConfigs)
			}
		})
	}
}

func TestPIP_NewPIP(t *testing.T) {
	t.Parallel()

	t.Run("new pip", func(t *testing.T) {
		p1 := &PIP{
			Store:        "../../testdata/unittest/pip2",
			StoreRecurse: false,
			PullConfigs:  "../../testdata/unittest/pip2/pull",
		}

		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		p2, err := p1.NewPIP(context.Background(), logger, models.CEDAR)
		require.NoError(t, err)
		require.NotNil(t, p2)
	})
}
