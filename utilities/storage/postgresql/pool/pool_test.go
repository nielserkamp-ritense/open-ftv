package pool

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPool(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			p, err := NewPool(ctx, tc.dsn, tc.opts...)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, p)
			} else {
				require.NoError(t, err)
				require.NotNil(t, p)

				p2, ok := p.(*Pool)
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
	t.Parallel()

	t.Run("pool ping", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectPing()

		p := &Pool{pool: mock}
		err = p.Ping(context.Background())
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPool_Ping_Fail(t *testing.T) {
	t.Parallel()

	t.Run("pool ping fail", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		cfg, err2 := pgxpool.ParseConfig("postgresql://localhost:5432/myDB")
		require.NoError(t, err2)
		require.NotNil(t, cfg)

		p := &Pool{pool: mock, cfg: cfg}
		err = p.Ping(context.Background())
		require.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPool_Begin(t *testing.T) {
	t.Parallel()

	t.Run("pool begin", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectBegin()
		mock.ExpectRollback()

		p := &Pool{pool: mock}
		tx, err2 := p.Begin(context.Background())
		require.NoError(t, err2)
		require.NotNil(t, tx)

		err2 = tx.Rollback(context.Background())
		require.NoError(t, err2)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPool_Close(t *testing.T) {
	t.Parallel()

	t.Run("pool close", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectClose()

		p := &Pool{pool: mock}
		p.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestErrors2(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("test error")
	cfg := &pgxpool.Config{ConnConfig: &pgx.ConnConfig{Config: pgconn.Config{Host: "localhost", Port: 5432, Database: "myDB"}}}
	p := &Pool{cfg: cfg}
	q := "INSERT INTO kv (key, index, value) VALUES ($1,$2,$3)"
	params := []any{"key1", 9, "hello world", true}

	id := "localhost:5432/myDB"
	qp := `[]interface {}{"key1", 9, "hello world", true}`

	testCases := []struct {
		name      string
		do        error
		contains1 string
		contains2 string
		contains3 string
		contains4 string
	}{
		{
			name:      "pool - dsn",
			do:        dsnError(err),
			contains1: "dsn parsing",
			contains2: "",
		},
		{
			name:      "pool - connection",
			do:        p.connectionError(err),
			contains1: "connection",
			contains2: id,
		},
		{
			name:      "pool - ping",
			do:        p.pingError(err),
			contains1: "ping",
			contains2: id,
		},
		{
			name:      "pool - tx",
			do:        p.TxError(err),
			contains1: "begin transaction",
			contains2: id,
		},
		{
			name:      "pool - query",
			do:        p.QueryError(err, q, params...),
			contains1: "query",
			contains2: id,
			contains3: q,
			contains4: qp,
		},
		{
			name:      "pool - scan",
			do:        p.ScanError(err, q, params...),
			contains1: "row scan",
			contains2: id,
			contains3: q,
			contains4: qp,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.do
			require.NotNil(t, got)
			assert.True(t, errors.Is(got, err))

			s := got.Error()
			assert.True(t, strings.Contains(s, tc.contains1))

			if tc.contains2 != "" {
				assert.True(t, strings.Contains(s, tc.contains2))
			}
			if tc.contains3 != "" {
				assert.True(t, strings.Contains(s, tc.contains3))
			}
			if tc.contains4 != "" {
				assert.True(t, strings.Contains(s, tc.contains4))
			}
		})
	}
}

func TestNewWithPooler(t *testing.T) {
	t.Parallel()

	t.Run("new with pooler", func(t *testing.T) {
		t.Parallel()

		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		dsn := "postgresql://localhost:5432/myDB"
		p, err2 := NewWithPooler(context.Background(), dsn, mock, WithMaxLifetime(time.Second))
		require.NoError(t, err2)
		require.NotNil(t, p)

		dsn = "oopsie"
		p, err2 = NewWithPooler(context.Background(), dsn, mock, WithMaxLifetime(time.Second))
		require.Error(t, err2)
		require.Nil(t, p)
	})
}
