package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
	"gitlab.com/gjuyn/go-config/config-ext/yaml"
)

func TestCerbos(t *testing.T) {
	t.Parallel()

	const goodCfg = `
cerbos:
  address: "https://localhost:1234"
  ca: "/etc/ssl/certs/ca.crt"
  admin:
    address: "https://localhost:1234/admin"
    user: "mickey"
    password: "mouse"
`

	dir := t.TempDir()

	testCases := []struct {
		name        string
		file        string
		data        string
		args        []string
		wantErr     bool
		wantAddress string
		wantAdmin   string
		wantCA      string
		wantUser    string
		wantPswd    string
	}{
		{
			name:    "bad input",
			file:    "cerbos1.yaml",
			data:    "\001\002",
			wantErr: true,
		},
		{
			name:        "good file",
			file:        "cerbos2.yaml",
			data:        goodCfg,
			wantAddress: "https://localhost:1234",
			wantAdmin:   "https://localhost:1234/admin",
			wantCA:      "/etc/ssl/certs/ca.crt",
			wantUser:    "mickey",
			wantPswd:    "mouse",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := dir + tc.file
			err := os.WriteFile(file, []byte(tc.data), 0644)
			require.NoError(t, err)

			cfg := &Cerbos{}

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
				assert.Equal(t, tc.wantAddress, cfg.Address)
				assert.Equal(t, tc.wantAdmin, cfg.AdminAddress)
				assert.Equal(t, tc.wantCA, cfg.CA)
				assert.Equal(t, tc.wantUser, cfg.User)
				assert.Equal(t, tc.wantPswd, cfg.Pswd)
			}
		})
	}
}

func TestCerbos_Sanitized(t *testing.T) {
	t.Parallel()

	t.Run("sanitize cerbos", func(t *testing.T) {
		c := &Cerbos{
			Address:      "adres",
			AdminAddress: "admin",
			CA:           "ca",
			User:         "user",
			Pswd:         "password",
		}

		s := c.Sanitized()

		assert.Equal(t, "adres", s.Address)
		assert.Equal(t, "admin", s.AdminAddress)
		assert.Equal(t, "ca", s.CA)
		assert.Empty(t, s.User)
		assert.Empty(t, s.Pswd)
	})
}
