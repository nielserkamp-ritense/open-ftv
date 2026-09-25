package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
)

func TestPDP(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		args        []string
		wantErr     bool
		wantMapping string
	}{
		{
			name:    "bad retries",
			args:    []string{"--bundle-retries=x"},
			wantErr: true,
		},
		{
			name:        "good args",
			args:        []string{"--request-mappings=a,b,c"},
			wantMapping: "a,b,c",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &PDP{}

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
				assert.Equal(t, tc.wantMapping, cfg.RequestMappings)
			}
		})
	}
}
