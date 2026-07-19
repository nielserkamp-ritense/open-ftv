package pip

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

func newMockPIP(t *testing.T, ctx context.Context) (pgxmock.PgxPoolIface, *PostgresDB) {
	t.Helper()

	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	require.NotNil(t, mock)

	pool, err2 := postgresql.NewWithPool(ctx, "postgres://localhost:5432/myDB", time.Minute, 3, mock)
	require.NoError(t, err2)
	require.NotNil(t, pool)

	return mock, NewPostgresWithPool(pool)
}

func TestPostgresDB_DeleteAttribute(t *testing.T) {
	t.Parallel()

	upd := time.Date(2025, 8, 13, 12, 10, 8, 123456789, time.UTC)
	lastIndex := timeToLastIndex(upd)
	ts := timeFromLastIndex(lastIndex)

	testCases := []struct {
		name    string
		count   int64
		dbErr   bool
		wantErr bool
	}{
		{name: "delete", count: 1},
		{name: "no rows affected", count: 0, wantErr: true},
		{name: "force error", dbErr: true, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mock, db := newMockPIP(t, ctx)

			prev := models.NewAttribute("attr1", "value")

			mock.ExpectBegin()
			mock.ExpectExec("SELECT set_config('openftv.user', $1, true)").
				WithArgs("*SYSTEM*").
				WillReturnResult(pgxmock.NewResult("SELECT", 1))

			exp := mock.ExpectExec(`DELETE FROM attribute WHERE key=$1 AND updated=$2`).
				WithArgs(prev.Key(), ts)
			if tc.dbErr {
				exp.WillReturnError(errors.New("test error"))
				mock.ExpectRollback()
			} else {
				exp.WillReturnResult(pgxmock.NewResult("DELETE", tc.count))
				mock.ExpectCommit()
			}

			got, err := db.DeleteAttribute(ctx, identity.NewSystemPrincipal(), prev, lastIndex)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, prev, got)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresDB_DeleteEntity(t *testing.T) {
	t.Parallel()

	upd := time.Date(2025, 8, 13, 12, 10, 8, 123456789, time.UTC)
	lastIndex := timeToLastIndex(upd)
	ts := timeFromLastIndex(lastIndex)

	testCases := []struct {
		name    string
		count   int64
		dbErr   bool
		wantErr bool
	}{
		{name: "delete", count: 1},
		{name: "no rows affected", count: 0, wantErr: true},
		{name: "force error", dbErr: true, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mock, db := newMockPIP(t, ctx)

			prev := models.NewEntity("user", "u1", nil)

			mock.ExpectBegin()
			mock.ExpectExec("SELECT set_config('openftv.user', $1, true)").
				WithArgs("*SYSTEM*").
				WillReturnResult(pgxmock.NewResult("SELECT", 1))

			exp := mock.ExpectExec(`DELETE FROM entity WHERE type=$1 AND id=$2 AND updated=$3`).
				WithArgs(prev.Type(), prev.ID(), ts)
			if tc.dbErr {
				exp.WillReturnError(errors.New("test error"))
				mock.ExpectRollback()
			} else {
				exp.WillReturnResult(pgxmock.NewResult("DELETE", tc.count))
				mock.ExpectCommit()
			}

			got, err := db.DeleteEntity(ctx, identity.NewSystemPrincipal(), prev, lastIndex)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, prev, got)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
