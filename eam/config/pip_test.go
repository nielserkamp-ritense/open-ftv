package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
)

func TestPIP(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		args        []string
		wantErr     bool
		wantStore   string
		wantRecurse bool
		wantPullCfg string
	}{
		{
			name:    "bad recurse",
			args:    []string{"--pip-store-recurse=x"},
			wantErr: true,
		},
		{
			name:        "good args",
			args:        []string{"--pip-store=/etc/pip", "--pip-store-recurse=true", "--pip-pull-configs=/etc/pip/pull"},
			wantStore:   "/etc/pip",
			wantRecurse: true,
			wantPullCfg: "/etc/pip/pull",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &PIP{}

			opts := []config.Option{
				config.WithArgs(tc.args...),
				config.EnvironmentPrefix("CFG_"),
				config.AppName("test 1.0"),
				config.NoHelpOnError(),
			}

			err := config.LoadConfig(cfg, opts...)
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
