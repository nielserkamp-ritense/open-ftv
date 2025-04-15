package postgresql

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("new handler", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectClose()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}

		db, err3 := New(p, "myTable")
		require.NoError(t, err3)
		require.NotNil(t, db)

		db2, ok := db.(*pgDB)
		require.True(t, ok)
		require.NotNil(t, db2)

		assert.Equal(t, p, db2.pool)
		assert.Equal(t, "myTable", db2.table)

		_, err = db.Watch(nil, "", nil)
		require.Error(t, err)

		_, err = db.WatchTree(nil, "", nil)
		require.Error(t, err)

		err = db.DeleteTree(nil, "")
		require.Error(t, err)

		_, err = db.NewLock(nil, "", nil)
		require.Error(t, err)

		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
