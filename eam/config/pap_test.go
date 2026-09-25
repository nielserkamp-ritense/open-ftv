package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
)

func TestPAP(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		args         []string
		wantErr      bool
		wantLanguage string
		wantStore    string
		wantRecurse  bool
	}{
		{
			name:    "bad recurse",
			args:    []string{"--policies-store-recurse=x"},
			wantErr: true,
		},
		{
			name:         "good args",
			args:         []string{"--policies-language=REGO", "--policies-store=/etc/opa", "--policies-store-recurse=true"},
			wantLanguage: "REGO",
			wantStore:    "/etc/opa",
			wantRecurse:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &PAP{}

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
				assert.Equal(t, tc.wantLanguage, cfg.Language)
				assert.Equal(t, tc.wantStore, cfg.Store)
				assert.Equal(t, tc.wantRecurse, cfg.StoreRecurse)
			}
		})
	}
}
