package postgresql

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB_Get(t *testing.T) {
	t.Parallel()

	t.Run("get", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT key, index, value FROM "myTable" WHERE key = $1`).
			WithArgs("xyz").
			WillReturnRows(mock.NewRows([]string{"key", "index", "value"}).
				AddRow("xyz", 1, []byte("value"))).
			RowsWillBeClosed()
		mock.ExpectCommit()
		mock.ExpectClose()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		rec, err4 := db.Get(context.Background(), "xyz", nil)
		require.NoError(t, err4)
		require.NotNil(t, rec)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_Get_FailTx(t *testing.T) {
	t.Parallel()

	t.Run("get - fail tx", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin().WillReturnError(errors.New("fail"))
		mock.ExpectClose()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		rec, err4 := db.Get(context.Background(), "xyz", nil)
		require.Error(t, err4)
		require.Nil(t, rec)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_Get_FailQuery(t *testing.T) {
	t.Parallel()

	t.Run("get - fail query", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT key, index, value FROM "myTable" WHERE key = $1`).
			WithArgs("xyz").
			WillReturnError(errors.New("fail"))
		mock.ExpectRollback()
		mock.ExpectClose()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		rec, err4 := db.Get(context.Background(), "xyz", nil)
		require.Error(t, err4)
		require.Nil(t, rec)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_Get_NotFound(t *testing.T) {
	t.Parallel()

	t.Run("get - not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT key, index, value FROM "myTable" WHERE key = $1`).
			WithArgs("xyz").
			WillReturnRows(mock.NewRows([]string{"key", "index", "value"}).
				RowError(0, errors.New("fail"))).
			RowsWillBeClosed()
		mock.ExpectRollback()
		mock.ExpectClose()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		rec, err4 := db.Get(context.Background(), "xyz", nil)
		require.Error(t, err4)
		require.Nil(t, rec)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_Get_FailScan(t *testing.T) {
	t.Parallel()

	t.Run("get - fail scan", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT key, index, value FROM "myTable" WHERE key = $1`).
			WithArgs("xyz").
			WillReturnRows(mock.NewRows([]string{"key", "value", "ok"}).
				AddRow("x", "y", true)).
			RowsWillBeClosed()
		mock.ExpectRollback()
		mock.ExpectClose()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		rec, err4 := db.Get(context.Background(), "xyz", nil)
		require.Error(t, err4)
		require.Nil(t, rec)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_List(t *testing.T) {
	t.Parallel()

	t.Run("list", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT key, index, value FROM "myTable" WHERE key LIKE $1`).
			WithArgs("a%").
			WillReturnRows(mock.NewRows([]string{"key", "index", "value"}).
				AddRow("axz", 1, []byte("value")).
				AddRow("abc", 2, []byte("some"))).
			RowsWillBeClosed()
		mock.ExpectCommit()
		mock.ExpectClose()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		rec, err4 := db.List(context.Background(), "a", nil)
		require.NoError(t, err4)
		require.NotNil(t, rec)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_List_FailTx(t *testing.T) {
	t.Parallel()

	t.Run("list - fail tx", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin().WillReturnError(errors.New("fail"))
		mock.ExpectClose()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		rec, err4 := db.List(context.Background(), "a", nil)
		require.Error(t, err4)
		require.Nil(t, rec)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_List_FailQuery(t *testing.T) {
	t.Parallel()

	t.Run("list - fail query", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT key, index, value FROM "myTable" WHERE key LIKE $1`).
			WithArgs("a%").
			WillReturnError(errors.New("fail"))
		mock.ExpectRollback()
		mock.ExpectClose()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		rec, err4 := db.List(context.Background(), "a", nil)
		require.Error(t, err4)
		require.Nil(t, rec)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_List_Empty(t *testing.T) {
	t.Parallel()

	t.Run("list - empty", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT key, index, value FROM "myTable" WHERE key LIKE $1`).
			WithArgs("a%").
			WillReturnRows(pgxmock.NewRows([]string{"key", "index", "value"}).
				RowError(0, errors.New("fail"))).
			RowsWillBeClosed()
		mock.ExpectRollback()
		mock.ExpectClose()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		rec, err4 := db.List(context.Background(), "a", nil)
		require.Error(t, err4)
		require.Nil(t, rec)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_List_FailScan(t *testing.T) {
	t.Run("list - fail scan", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT key, index, value FROM "myTable" WHERE key LIKE $1`).
			WithArgs("a%").
			WillReturnRows(pgxmock.NewRows([]string{"key", "value", "ok"}).
				AddRow("x", "y", true)).
			RowsWillBeClosed()
		mock.ExpectRollback()
		mock.ExpectClose()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		rec, err4 := db.List(context.Background(), "a", nil)
		require.Error(t, err4)
		require.Nil(t, rec)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_Exists(t *testing.T) {
	t.Parallel()

	t.Run("exists", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT key, index, value FROM "myTable" WHERE key = $1`).
			WithArgs("xyz").
			WillReturnRows(mock.NewRows([]string{"key", "index", "value"}).
				AddRow("xyz", 1, []byte("value"))).
			RowsWillBeClosed()
		mock.ExpectCommit()
		mock.ExpectClose()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		ok, err4 := db.Exists(context.Background(), "xyz", nil)
		require.NoError(t, err4)
		require.True(t, ok)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_Exists_Fail(t *testing.T) {
	t.Parallel()

	t.Run("exists", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin().WillReturnError(errors.New("fail"))
		mock.ExpectClose()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		ok, err4 := db.Exists(context.Background(), "xyz", nil)
		require.Error(t, err4)
		require.False(t, ok)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
