package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
)

func TestServer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		args     []string
		wantErr  bool
		wantHost string
		wantPort uint16
	}{
		{
			name:    "bad port",
			args:    []string{"--port=x"},
			wantErr: true,
		},
		{
			name:     "good args",
			args:     []string{"--host=127.0.0.1", "--port=8080"},
			wantHost: "127.0.0.1",
			wantPort: 8080,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &Server{}

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
				assert.Equal(t, tc.wantHost, cfg.Host)
				assert.Equal(t, tc.wantPort, cfg.Port)
			}
		})
	}
}
