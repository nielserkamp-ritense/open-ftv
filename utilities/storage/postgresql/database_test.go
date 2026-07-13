package postgresql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgres_Exec(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		q           string
		params      []any
		errBegin    bool
		errExec     bool
		errRollback bool
		errCommit   bool
		wantOP      string
		wantCount   int64
	}{
		{
			name:     "begin fail",
			q:        `not important`,
			errBegin: true,
		},
		{
			name:    "exec fail",
			q:       `THIS IS BAD SQL`,
			errExec: true,
		},
		{
			name:        "exec & rollback fail",
			q:           `THIS IS BAD SQL`,
			errExec:     true,
			errRollback: true,
		},
		{
			name:      "commit fail",
			q:         `CREATE TABLE "test" (id INTEGER PRIMARY KEY, title VARCHAR(40))`,
			errCommit: true,
		},
		{
			name: "create",
			q:    `CREATE TABLE "test" (id INTEGER PRIMARY KEY, title VARCHAR(40))`,
		},
		{
			name:      "insert",
			q:         `INSERT INTO "test" (id, title) VALUES ($1, $2)`,
			params:    []any{1, "hello world"},
			wantOP:    "INSERT",
			wantCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			u := "*SYSTEM*"

			mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer mock.Close()

			exp1 := mock.ExpectBegin()
			if tc.errBegin {
				exp1.WillReturnError(errors.New("begin"))
			}

			mock.ExpectExec("SELECT set_config('openftv.user', $1, true)").WithArgs(u).WillReturnResult(pgxmock.NewResult("SELECT", 1))

			exp2 := mock.ExpectExec(tc.q)
			if len(tc.params) > 0 {
				exp2 = exp2.WithArgs(tc.params...)
			}
			if tc.errExec {
				exp2.WillReturnError(errors.New("exec"))

				exp3 := mock.ExpectRollback()
				if tc.errRollback {
					exp3.WillReturnError(errors.New("rollback"))
				}
			} else {
				exp2.WillReturnResult(pgxmock.NewResult(tc.wantOP, tc.wantCount))

				exp3 := mock.ExpectCommit()
				if tc.errCommit {
					exp3.WillReturnError(errors.New("commit"))
				}
			}

			mock.ExpectClose()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			db := &Postgres{dsn: "x", maxLife: time.Minute, maxConn: 5, pool: mock}

			count, err2 := db.Exec(ctx, tc.q, tc.params)

			if tc.errBegin || tc.errExec || tc.errRollback || tc.errCommit {
				require.Error(t, err2)
				assert.Zero(t, count)
			} else {
				require.NoError(t, err2)
				assert.Equal(t, tc.wantCount, count)
			}
		})
	}
}

func TestPostgres_Query(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		q           string
		params      []any
		rows        *pgxmock.Rows
		errBegin    bool
		errExec     bool
		errScan     bool
		errRollback bool
		errCommit   bool
		rowBreak    int64
		wantCount   int64
	}{
		{
			name:     "begin fail",
			q:        `not important`,
			errBegin: true,
		},
		{
			name:    "exec fail",
			q:       `SELECT iets`,
			errExec: true,
		},
		{
			name:        "query & rollback fail",
			q:           `SELECT iets`,
			errExec:     true,
			errRollback: true,
		},
		{
			name: "row scan fail",
			q:    `SELECT * FROM "test"`,
			rows: pgxmock.NewRows([]string{"id", "title"}).
				AddRows([]any{1, "hello world"}, []any{2, "oops"}).
				RowError(1, errors.New("row")),
			errScan:   true,
			wantCount: 1,
		},
		{
			name:      "commit fail",
			q:         `SELECT * FROM "test"`,
			rows:      pgxmock.NewRows([]string{"id", "title"}),
			errCommit: true,
		},
		{
			name: "select",
			q:    `SELECT * FROM "test"`,
			rows: pgxmock.NewRows([]string{"id", "title"}).
				AddRows([]any{1, "yes"}, []any{2, "no"}, []any{3, "maybe"}, []any{4, "what if"}, []any{5, "whatever"}),
			wantCount: 5,
		},
		{
			name: "select with row break",
			q:    `SELECT * FROM "test"`,
			rows: pgxmock.NewRows([]string{"id", "title"}).
				AddRows([]any{1, "yes"}, []any{2, "no"}, []any{3, "maybe"}, []any{4, "what if"}, []any{5, "whatever"}),
			rowBreak:  3,
			wantCount: 3,
		},
		{
			name:      "select with params",
			q:         `SELECT * FROM "test" WHERE id = $1`,
			params:    []any{1},
			rows:      pgxmock.NewRows([]string{"id", "title"}).AddRows([]any{1, "yes"}),
			wantCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			require.NoError(t, err)
			defer mock.Close()

			exp1 := mock.ExpectBegin()
			if tc.errBegin {
				exp1.WillReturnError(errors.New("begin"))
			}

			exp2 := mock.ExpectQuery(tc.q)
			if len(tc.params) > 0 {
				exp2 = exp2.WithArgs(tc.params...)
			}
			if tc.errExec {
				exp2.WillReturnError(errors.New("exec"))

				exp3 := mock.ExpectRollback()
				if tc.errRollback {
					exp3.WillReturnError(errors.New("rollback"))
				}
			} else {
				exp2.WillReturnRows(tc.rows)

				exp3 := mock.ExpectCommit()
				if tc.errCommit {
					exp3.WillReturnError(errors.New("commit"))
				}
			}

			mock.ExpectClose()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			db := &Postgres{dsn: "x", maxLife: time.Minute, maxConn: 5, pool: mock}

			var count int64
			err2 := db.Query(ctx, tc.q, tc.params, func(values []any) bool {
				count++
				return tc.rowBreak == 0 || count != tc.rowBreak
			})

			if tc.errBegin || tc.errExec || tc.errScan || tc.errRollback || tc.errCommit {
				require.Error(t, err2)
			} else {
				require.NoError(t, err2)
			}

			assert.Equal(t, tc.wantCount, count)
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		dsn     string
		maxLife time.Duration
		maxConn int32
		wantErr bool
	}{
		{
			name:    "bad dsn",
			dsn:     "oops",
			maxLife: time.Minute,
			maxConn: 3,
			wantErr: true,
		},
		{
			name:    "good",
			dsn:     "postgres://localhost:5432/myDB",
			maxLife: time.Minute,
			maxConn: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			p, err := New(ctx, tc.dsn, tc.maxLife, tc.maxConn)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, p)
			} else {
				require.NoError(t, err)
				require.NotNil(t, p)

				assert.Equal(t, tc.dsn, p.dsn)
				assert.Equal(t, tc.maxLife, p.maxLife)
				assert.Equal(t, tc.maxConn, p.maxConn)
				assert.NotNil(t, p.pool)

				// allow a graceful shutdown.
				cancel()
				time.Sleep(time.Millisecond * 25)
			}
		})
	}
}
