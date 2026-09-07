package config

import (
	"context"
	"os"
	"testing"
	"time"

	"gitlab.com/gjuyn/go-config/config-ext/yaml"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"
)

func TestPersist(t *testing.T) {
	t.Parallel()

	const badRecurse = `
persist:
  type: "postgres"
  prefix: "base"
  timeout: "10s"
  postgres: 
    url: "postgres://me@localhost:5432/db"
    table: "table"
    connection:
      ttl: 25s
      max: 15
`

	const goodCfg = `
persist:
  type: "etcd"
  addresses: "localhost:1111"
  prefix: "/pip/persist"
  timeout: "15s"
  etcd:
    sync: "5m"
    user: "mickey"
    password: "mouse"
`
	dir := t.TempDir()

	testCases := []struct {
		name            string
		file            string
		data            string
		args            []string
		wantErr         bool
		wantType        string
		wantAddresses   string
		wantBase        string
		wantTimout      time.Duration
		wantEtcdSync    time.Duration
		wantEtcdUser    string
		wantEtcdPswd    string
		wantConsulToken string
		wantConsulNS    string
		wantPgURL       string
		wantPgTable     string
		wantPgMaxLife   time.Duration
		wantPgMaxConn   int32
	}{
		{
			name:    "bad input",
			file:    "recurse1.yaml",
			data:    "%not yaml%",
			wantErr: true,
		},
		{
			name:          "bad recurse",
			file:          "recurse2.yaml",
			data:          badRecurse,
			wantType:      "postgres",
			wantBase:      "base",
			wantTimout:    10 * time.Second,
			wantPgURL:     "postgres://me@localhost:5432/db",
			wantPgTable:   "table",
			wantPgMaxLife: 25 * time.Second,
			wantPgMaxConn: 15,
		},
		{
			name:          "good file",
			file:          "recurse3.yaml",
			data:          goodCfg,
			wantType:      "etcd",
			wantAddresses: "localhost:1111",
			wantBase:      "/pip/persist",
			wantTimout:    15 * time.Second,
			wantEtcdSync:  5 * time.Minute,
			wantEtcdUser:  "mickey",
			wantEtcdPswd:  "mouse",
			wantPgMaxLife: 5 * time.Minute,
			wantPgMaxConn: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := dir + tc.file
			err := os.WriteFile(file, []byte(tc.data), 0644)

			cfg := &Persist{}

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
				assert.Equal(t, tc.wantType, cfg.Type)
				assert.Equal(t, tc.wantAddresses, cfg.Addresses)
				assert.Equal(t, tc.wantBase, cfg.Base)
				assert.Equal(t, tc.wantTimout, cfg.Timeout)
				assert.Equal(t, tc.wantEtcdSync, cfg.EtcdSync)
				assert.Equal(t, tc.wantEtcdUser, cfg.EtcdUser)
				assert.Equal(t, tc.wantEtcdPswd, cfg.EtcdPswd)
				assert.Equal(t, tc.wantConsulToken, cfg.ConsulToken)
				assert.Equal(t, tc.wantConsulNS, cfg.ConsulNamespace)
				assert.Equal(t, tc.wantPgURL, cfg.PgURL)
				assert.Equal(t, tc.wantPgTable, cfg.PgTable)
				assert.Equal(t, tc.wantPgMaxLife, cfg.PgMaxLife)
				assert.Equal(t, tc.wantPgMaxConn, cfg.PgMaxConn)
			}
		})
	}
}

func TestPersist_Sanitized(t *testing.T) {
	t.Parallel()

	t.Run("sanitize persist", func(t *testing.T) {
		p := &Persist{
			Type:            "type",
			Addresses:       "address",
			Base:            "base",
			Timeout:         time.Second,
			EtcdSync:        time.Second,
			EtcdUser:        "user",
			EtcdPswd:        "pswd",
			ConsulToken:     "token",
			ConsulNamespace: "namespace",
			PgURL:           "url",
			PgTable:         "table",
			PgMaxLife:       time.Minute,
			PgMaxConn:       15,
		}

		s := p.Sanitized()

		assert.Equal(t, "type", s.Type)
		assert.Equal(t, "address", s.Addresses)
		assert.Equal(t, "base", s.Base)
		assert.Equal(t, time.Second, s.Timeout)
		assert.Equal(t, time.Second, s.EtcdSync)
		assert.Empty(t, s.EtcdUser)
		assert.Empty(t, s.EtcdPswd)
		assert.Empty(t, s.ConsulToken)
		assert.Equal(t, "namespace", s.ConsulNamespace)
		assert.Empty(t, s.PgURL)
		assert.Equal(t, "table", s.PgTable)
		assert.Equal(t, time.Minute, s.PgMaxLife)
		assert.Equal(t, int32(15), s.PgMaxConn)
	})
}

func TestNewStore(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		cfg  *Persist
		fail bool
		want bool
	}{
		{
			name: "no type",
			cfg:  &Persist{},
			fail: true,
		},
		{
			name: "bad type",
			cfg: &Persist{
				Type:      "xyz",
				Addresses: "https://bad.host.localhost",
			},
			fail: true,
		},
		{
			name: "memory",
			cfg: &Persist{
				Type: "memory",
			},
			fail: true,
		},
		{
			name: "postgres",
			cfg: &Persist{
				Type:      "pg",
				PgURL:     "postgres://localhost:5432/myDB",
				PgTable:   "myTable",
				PgMaxLife: 60 * time.Second,
				PgMaxConn: 10,
			},
			want: true,
		},
		{
			name: "postgres fail",
			cfg: &Persist{
				Type:  "pg",
				PgURL: "postgres://localhost:5432/myDB",
			},
			fail: true,
		},
		{
			name: "etcd",
			cfg: &Persist{
				Type:      "etcd",
				Addresses: "https://bad.host.localhost",
			},
			want: true,
		},
		{
			name: "consul",
			cfg: &Persist{
				Type:      "consul",
				Addresses: "https://bad.host.localhost",
			},
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			s, err := tc.cfg.NewStore(ctx)
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
