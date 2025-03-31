package postgresql

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	err := fmt.Errorf("test error")
	cfg := &pgxpool.Config{ConnConfig: &pgx.ConnConfig{Config: pgconn.Config{Host: "localhost", Port: 5432, Database: "myDB"}}}
	p := &pool{cfg: cfg}
	db1 := &pgDB{pool: p}
	db2 := &pgDB{}
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
			do:        p.txError(err),
			contains1: "begin transaction",
			contains2: id,
		},
		{
			name:      "pool - query",
			do:        p.queryError(err, q, params...),
			contains1: "query",
			contains2: id,
			contains3: q,
			contains4: qp,
		},
		{
			name:      "pool - scan",
			do:        p.scanError(err, q, params...),
			contains1: "row scan",
			contains2: id,
			contains3: q,
			contains4: qp,
		},
		{
			name:      "db with pool - tx",
			do:        db1.txError(err),
			contains1: "begin transaction",
			contains2: id,
		},
		{
			name:      "db with pool - query",
			do:        db1.queryError(err, q, params...),
			contains1: "query",
			contains2: id,
			contains3: q,
			contains4: qp,
		},
		{
			name:      "db with pool - scan",
			do:        db1.scanError(err, q, params...),
			contains1: "row scan",
			contains2: id,
			contains3: q,
			contains4: qp,
		},
		{
			name:      "db without pool - tx",
			do:        db2.txError(err),
			contains1: "begin transaction",
		},
		{
			name:      "db without pool - query",
			do:        db2.queryError(err, q, params...),
			contains1: "query",
			contains3: q,
			contains4: qp,
		},
		{
			name:      "db without pool - scan",
			do:        db2.scanError(err, q, params...),
			contains1: "row scan",
			contains3: q,
			contains4: qp,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
