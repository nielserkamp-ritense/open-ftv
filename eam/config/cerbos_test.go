package config

import (
	"flag"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCerbos(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		args        []string
		wantErr     bool
		wantAddress string
		wantAdmin   string
		wantCA      string
		wantUser    string
		wantPswd    string
	}{
		{
			name:    "unknown flag",
			args:    []string{"--cerbos-unknown=x"},
			wantErr: true,
		},
		{
			name:        "good args",
			args:        []string{"--cerbos-address=https://localhost:1234", "--cerbos-admin=https://localhost:1234/admin", "--cerbos-ca=/etc/ssl/certs/ca.crt", "--cerbos-user=mickey", "--cerbos-pswd=mouse"},
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

			cfg := &Cerbos{}

			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(io.Discard)
			cfg.RegisterFlags(fs)

			err := fs.Parse(tc.args)
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
