package postgresql

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kvtools/valkeyrie/store"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB_Put(t *testing.T) {
	t.Parallel()

	t.Run("put", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "myTable" (key, index, value) VALUES ($1, 1, $2)
 ON CONFLICT (key) DO UPDATE SET index = myTable.index + 1, value = EXCLUDED.value`).
			WithArgs("xyz", []byte("blah")).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
		mock.ExpectCommit()
		mock.ExpectClose()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		err4 := db.Put(context.Background(), "xyz", []byte("blah"), nil)
		require.NoError(t, err4)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_Put_FailTx(t *testing.T) {
	t.Parallel()

	t.Run("put - fail tx", func(t *testing.T) {
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

		err4 := db.Put(context.Background(), "xyz", []byte("blah"), nil)
		require.Error(t, err4)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_Put_FailExec(t *testing.T) {
	t.Parallel()

	t.Run("put - fail exec", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "myTable" (key, index, value) VALUES ($1, 1, $2)
 ON CONFLICT (key) DO UPDATE SET index = myTable.index + 1, value = EXCLUDED.value`).
			WithArgs("xyz", []byte("blah")).
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

		err4 := db.Put(context.Background(), "xyz", []byte("blah"), nil)
		require.Error(t, err4)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_AtomicPut_Insert(t *testing.T) {
	t.Parallel()

	t.Run("atomic put insert", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "myTable" (key, index, value) VALUES ($1, 1, $2)`).
			WithArgs("xyz", []byte("blah")).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
		mock.ExpectCommit()
		mock.ExpectClose()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		ok, next, err4 := db.AtomicPut(context.Background(), "xyz", []byte("blah"), nil, nil)
		require.NoError(t, err4)
		require.True(t, ok)
		require.NotNil(t, next)
		assert.Equal(t, uint64(1), next.LastIndex)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_AtomicPut_Update(t *testing.T) {
	t.Parallel()

	t.Run("atomic put update", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "myTable" SET index = index + 1, value = $1 WHERE key = $2 AND index = $3`).
			WithArgs([]byte("blah"), "xyz", uint64(5)).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		mock.ExpectCommit()
		mock.ExpectClose()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		prev := &store.KVPair{Key: "xyz", Value: []byte("oops"), LastIndex: 5}

		ok, next, err4 := db.AtomicPut(context.Background(), "xyz", []byte("blah"), prev, nil)
		require.NoError(t, err4)
		require.True(t, ok)
		require.NotNil(t, next)
		assert.Equal(t, uint64(6), next.LastIndex)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_AtomicPut_FailTx(t *testing.T) {
	t.Run("atomic put - fail tx", func(t *testing.T) {
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

		ok, next, err4 := db.AtomicPut(context.Background(), "xyz", []byte("blah"), nil, nil)
		require.Error(t, err4)
		require.False(t, ok)
		require.Nil(t, next)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_AtomicPut_FailExec(t *testing.T) {
	t.Parallel()

	t.Run("atomic put - fail exec", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO "myTable" (key, index, value) VALUES ($1, 1, $2)`).
			WithArgs("xyz", []byte("blah")).
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

		ok, next, err4 := db.AtomicPut(context.Background(), "xyz", []byte("blah"), nil, nil)
		require.Error(t, err4)
		require.False(t, ok)
		require.Nil(t, next)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
