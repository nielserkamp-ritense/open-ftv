package server

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pip/config"
	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/config"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNewPIP(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		language string
		persist  config2.Persist
		wantErr  bool
	}{
		{
			name: "no language",
			persist: config2.Persist{
				Type:      "pg",
				Base:      "base",
				Timeout:   30 * time.Second,
				PgURL:     "postgres://localhost:5432/postgres?sslmode=disable",
				PgTable:   "table",
				PgMaxLife: 25 * time.Second,
				PgMaxConn: 10,
			},
		},
		{
			name:     "bad language",
			language: "swahili",
			persist: config2.Persist{
				Type:      "pg",
				Base:      "base",
				Timeout:   30 * time.Second,
				PgURL:     "postgres://localhost:5432/postgres?sslmode=disable",
				PgTable:   "table",
				PgMaxLife: 25 * time.Second,
				PgMaxConn: 10,
			},
		},
		{
			name:     "no type",
			language: "cedar",
			persist:  config2.Persist{},
		},
		{
			name:     "bad type",
			language: "cedar",
			persist: config2.Persist{
				Type:    "DB2",
				Base:    "base",
				Timeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name:     "bad persist parameters",
			language: "cedar",
			persist: config2.Persist{
				Type:    "pg",
				Base:    "base",
				Timeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name:     "all good",
			language: "cedar",
			persist: config2.Persist{
				Type:      "pg",
				Base:      "base",
				Timeout:   30 * time.Second,
				PgURL:     "postgres://localhost:5432/postgres?sslmode=disable",
				PgTable:   "table",
				PgMaxLife: 25 * time.Second,
				PgMaxConn: 10,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelWarn)
			logger := slog.New(h)

			s := &service{
				ctx: ctx,
				cfg: &config.Config{
					PAP: config2.PAP{
						Language: tc.language,
					},
					Persist: tc.persist,
				},
				logger: logger,
			}

			p, err := s.newPIP()
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, p)
			} else {
				require.NoError(t, err)
				require.NotNil(t, p)
			}
		})
	}
}
