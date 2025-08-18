package postgresql

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql/pool"
)

func TestErrors(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer mock.Close()

	dsn := "postgresql://localhost:5432/myDB"
	p, err2 := pool.NewWithPooler(context.Background(), dsn, mock)
	require.NoError(t, err2)

	db1 := &pgDB{pool: p}
	db2 := &pgDB{}
	q := "INSERT INTO kv (key, index, value) VALUES ($1,$2,$3)"
	params := []any{"key1", 9, "hello world", true}

	id := "localhost:5432/myDB"
	qp := `[]interface {}{"key1", 9, "hello world", true}`

	err = fmt.Errorf("test error")

	testCases := []struct {
		name      string
		do        error
		contains1 string
		contains2 string
		contains3 string
		contains4 string
	}{
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
