package persistence

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/manager/config"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name string
		cfg  *config.Config
		fail bool
		want bool
	}{
		{
			name: "no type",
			cfg:  &config.Config{},
			fail: true,
		},
		{
			name: "bad type",
			cfg: &config.Config{
				PersistType:      "xyz",
				PersistAddresses: "https://bad.host.localhost",
			},
			fail: true,
		},
		{
			name: "memory",
			cfg: &config.Config{
				PersistType: "memory",
			},
		},
		{
			name: "etcd",
			cfg: &config.Config{
				PersistType:      "etcd",
				PersistAddresses: "https://bad.host.localhost",
			},
			want: true,
		},
		{
			name: "consul",
			cfg: &config.Config{
				PersistType:      "consul",
				PersistAddresses: "https://bad.host.localhost",
			},
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			s, err := New(ctx, tc.cfg)
			if tc.fail {
				require.Error(t, err)
				require.Nil(t, s)
			} else {
				require.NoError(t, err)
				if tc.want {
					require.NotNil(t, s)
				} else {
					require.Nil(t, s)
				}
			}
		})
	}
}
