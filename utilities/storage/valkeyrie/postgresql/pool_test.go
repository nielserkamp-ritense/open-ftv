package postgresql

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPool(t *testing.T) {
	testCases := []struct {
		name        string
		dsn         string
		opts        []Option
		wantErr     bool
		wantMaxLife time.Duration
		wantMaxConn int32
	}{
		{
			name:    "bad dsn",
			dsn:     "xyz",
			wantErr: true,
		},
		{
			name:        "postgres url",
			dsn:         "postgres://localhost:5432/myDB?sslmode=disable",
			wantMaxLife: 5 * time.Minute,
			wantMaxConn: 100,
		},
		{
			name:        "key/value pair",
			dsn:         "dbname = myDB",
			wantMaxLife: 5 * time.Minute,
			wantMaxConn: 100,
		},
		{
			name:        "good dsn - conn max life",
			dsn:         "postgres://localhost:5432/myDB?sslmode=disable",
			opts:        []Option{WithMaxLifetime(30 * time.Minute)},
			wantMaxLife: 30 * time.Minute,
			wantMaxConn: 100,
		},
		{
			name:        "good dsn - max conn",
			dsn:         "postgres://localhost:5432/myDB?sslmode=disable",
			opts:        []Option{WithMaxConnections(25)},
			wantMaxLife: 5 * time.Minute,
			wantMaxConn: 25,
		},
		{
			name:    "good dsn - bad max conn",
			dsn:     "postgres://localhost:5432/myDB?sslmode=disable",
			opts:    []Option{WithMaxConnections(0)},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			p, err := NewPool(ctx, tc.dsn, tc.opts...)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, p)
			} else {
				require.NoError(t, err)
				require.NotNil(t, p)

				p2, ok := p.(*pool)
				require.True(t, ok)
				require.NotNil(t, p2)

				assert.Equal(t, tc.wantMaxLife, p2.cfg.MaxConnLifetime)
				assert.Equal(t, tc.wantMaxConn, p2.cfg.MaxConns)
			}

			time.Sleep(10 * time.Millisecond)
		})
	}
}

func TestPool_Ping(t *testing.T) {
	t.Run("pool ping", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectPing()

		p := &pool{pool: mock}
		err = p.Ping(context.Background())
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPool_Ping_Fail(t *testing.T) {
	t.Run("pool ping fail", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &pool{pool: mock, cfg: cfg}
		err = p.Ping(context.Background())
		require.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPool_Begin(t *testing.T) {
	t.Run("pool begin", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectRollback()

		p := &pool{pool: mock}
		tx, err2 := p.Begin(context.Background())
		require.NoError(t, err2)
		require.NotNil(t, tx)

		err2 = tx.Rollback(context.Background())
		require.NoError(t, err2)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPool_Close(t *testing.T) {
	t.Run("pool close", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectClose()

		p := &pool{pool: mock}
		p.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
