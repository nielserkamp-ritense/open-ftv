package postgresql

import (
	"context"
	"errors"
	"testing"

	"github.com/kvtools/valkeyrie/store"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql/pool"
)

func TestDB_Delete(t *testing.T) {
	t.Parallel()

	t.Run("delete", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "myTable" WHERE key = $1`).
			WithArgs("xyz").
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		mock.ExpectCommit()
		mock.ExpectClose()

		dsn := "postgresql://localhost:5432/myDB"
		p, err2 := pool.NewWithPooler(context.Background(), dsn, mock)
		require.NoError(t, err2)

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		err4 := db.Delete(context.Background(), "xyz")
		require.NoError(t, err4)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_Delete_FailTx(t *testing.T) {
	t.Parallel()

	t.Run("delete - fail tx", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin().WillReturnError(errors.New("fail"))
		mock.ExpectClose()

		dsn := "postgresql://localhost:5432/myDB"
		p, err2 := pool.NewWithPooler(context.Background(), dsn, mock)
		require.NoError(t, err2)

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		err4 := db.Delete(context.Background(), "xyz")
		require.Error(t, err4)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_Delete_FailExec(t *testing.T) {
	t.Parallel()

	t.Run("delete", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "myTable" WHERE key = $1`).
			WithArgs("xyz").
			WillReturnError(errors.New("fail"))
		mock.ExpectRollback()
		mock.ExpectClose()

		dsn := "postgresql://localhost:5432/myDB"
		p, err2 := pool.NewWithPooler(context.Background(), dsn, mock)
		require.NoError(t, err2)

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		err4 := db.Delete(context.Background(), "xyz")
		require.Error(t, err4)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_AtomicDelete(t *testing.T) {
	t.Parallel()

	t.Run("atomic delete", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "myTable" WHERE key = $1 AND index = $2`).
			WithArgs("xyz", uint64(7)).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		mock.ExpectCommit()
		mock.ExpectClose()

		dsn := "postgresql://localhost:5432/myDB"
		p, err2 := pool.NewWithPooler(context.Background(), dsn, mock)
		require.NoError(t, err2)

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		prev := &store.KVPair{
			Key:       "xyz",
			Value:     []byte("blah"),
			LastIndex: 7,
		}

		ok, err4 := db.AtomicDelete(context.Background(), "xyz", prev)
		require.NoError(t, err4)
		require.True(t, ok)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_AtomicDelete_FailTx(t *testing.T) {
	t.Parallel()

	t.Run("atomic delete - fail tx", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin().WillReturnError(errors.New("fail"))
		mock.ExpectClose()

		dsn := "postgresql://localhost:5432/myDB"
		p, err2 := pool.NewWithPooler(context.Background(), dsn, mock)
		require.NoError(t, err2)

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		prev := &store.KVPair{
			Key:       "xyz",
			Value:     []byte("blah"),
			LastIndex: 7,
		}

		ok, err4 := db.AtomicDelete(context.Background(), "xyz", prev)
		require.Error(t, err4)
		require.False(t, ok)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_AtomicDelete_FailExec(t *testing.T) {
	t.Parallel()

	t.Run("atomic delete", func(t *testing.T) {
		mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "myTable" WHERE key = $1 AND index = $2`).
			WithArgs("xyz", uint64(7)).
			WillReturnError(errors.New("fail"))
		mock.ExpectRollback()
		mock.ExpectClose()

		dsn := "postgresql://localhost:5432/myDB"
		p, err2 := pool.NewWithPooler(context.Background(), dsn, mock)
		require.NoError(t, err2)

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		prev := &store.KVPair{
			Key:       "xyz",
			Value:     []byte("blah"),
			LastIndex: 7,
		}

		ok, err4 := db.AtomicDelete(context.Background(), "xyz", prev)
		require.Error(t, err4)
		require.False(t, ok)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
