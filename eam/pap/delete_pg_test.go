package pap

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

func newMockTagDB(t *testing.T, ctx context.Context) (pgxmock.PgxPoolIface, *TagDB) {
	t.Helper()

	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	require.NotNil(t, mock)

	pool, err2 := postgresql.NewWithPool(ctx, "postgres://localhost:5432/myDB", time.Minute, 3, mock)
	require.NoError(t, err2)
	require.NotNil(t, pool)

	return mock, NewTagDBWithPool(pool)
}

func TestTagDB_DeleteTag(t *testing.T) {
	t.Parallel()

	upd := time.Date(2025, 8, 13, 12, 10, 8, 123456789, time.UTC)
	lastIndex := timeToLastIndex(upd)
	ts := timeFromLastIndex(lastIndex)

	testCases := []struct {
		name      string
		rows      int64
		wantErr   bool
		errSubstr string
	}{
		{name: "delete", rows: 1},
		{name: "stale version", rows: 0, wantErr: true, errSubstr: "tag concurrency conflict"},
		{name: "force error", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mock, db := newMockTagDB(t, ctx)

			prev := &oas.Tag{Id: "rvig"}

			mock.ExpectBegin()
			mock.ExpectExec("SELECT set_config('openftv.user', $1, true)").
				WithArgs("*SYSTEM*").
				WillReturnResult(pgxmock.NewResult("SELECT", 1))

			exp := mock.ExpectExec(`DELETE FROM tag WHERE tag=$1 AND updated = $2`).
				WithArgs(prev.Id, ts)
			if tc.name == "force error" {
				exp.WillReturnError(errors.New("test error"))
				mock.ExpectRollback()
			} else {
				exp.WillReturnResult(pgxmock.NewResult("DELETE", tc.rows))
				mock.ExpectCommit()
			}

			got, err := db.DeleteTag(ctx, prev, lastIndex)
			if tc.wantErr {
				require.Error(t, err)
				if tc.errSubstr != "" {
					assert.ErrorContains(t, err, tc.errSubstr)
				}
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, prev, got)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTagDB_ReadTag(t *testing.T) {
	t.Parallel()

	const sql = `SELECT tag,title,description,created,created_by,updated,updated_by FROM tag WHERE tag=$1`
	cols := []string{"tag", "title", "description", "created", "created_by", "updated", "updated_by"}

	created := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	updated := time.Date(2025, 8, 13, 12, 10, 8, 123456789, time.UTC)

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mock, db := newMockTagDB(t, ctx)

		mock.ExpectBegin()
		mock.ExpectQuery(sql).
			WithArgs("rvig").
			WillReturnRows(pgxmock.NewRows(cols).
				AddRow("rvig", "RvIG policies", "", created, "creator", updated, "updater"))
		mock.ExpectCommit()

		got, ix, err := db.ReadTag(ctx, "rvig")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "rvig", got.Id)
		assert.Equal(t, timeToLastIndex(updated), ix)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	// Regression: a missing tag must not panic (nil-pointer dereference) but
	// return (nil, 0, nil) like ReadPolicy/ReadAttribute do.
	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mock, db := newMockTagDB(t, ctx)

		mock.ExpectBegin()
		mock.ExpectQuery(sql).
			WithArgs("missing").
			WillReturnRows(pgxmock.NewRows(cols))
		mock.ExpectCommit()

		got, ix, err := db.ReadTag(ctx, "missing")
		require.NoError(t, err)
		assert.Nil(t, got)
		assert.Zero(t, ix)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("force error", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mock, db := newMockTagDB(t, ctx)

		mock.ExpectBegin()
		mock.ExpectQuery(sql).
			WithArgs("rvig").
			WillReturnError(errors.New("test error"))
		mock.ExpectRollback()

		got, ix, err := db.ReadTag(ctx, "rvig")
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Zero(t, ix)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
