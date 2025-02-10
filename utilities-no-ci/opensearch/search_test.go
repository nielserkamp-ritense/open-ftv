package opensearch

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearcher_SearchBySQL(t *testing.T) {
	t.Run("search by sql", func(t *testing.T) {
		l, err := NewLogger(user, pswd, endpoints)
		require.NoError(t, err)
		require.NotNil(t, l)

		index := newID()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = l.CreateIndex(ctx, index, 1, 1)
		require.NoError(t, err)

		defer func() {
			err = l.DeleteIndexes(ctx, index)
			assert.NoError(t, err)
		}()

		err = l.LogBulk(ctx, true,
			LogRecord{
				Index: index,
				Data:  map[string]any{"hello": "world", "int": 1, "bool": true},
			},
			LogRecord{
				Index: index,
				Data:  map[string]any{"hello": "world2", "int": 2, "bool": false},
			},
			LogRecord{
				Index: index,
				Data:  map[string]any{"hello": "world3", "int": 3, "bool": true},
			},
		)
		require.NoError(t, err)

		s, err2 := NewSearcher(user, pswd, endpoints)
		require.NoError(t, err2)
		require.NotNil(t, s)

		got, err3 := s.SearchBySQL(ctx, index, fmt.Sprintf("select 'hello' from %s;", index), 999)
		require.NoError(t, err3)
		require.NotNil(t, got)
		assert.Equal(t, int64(3), got.Hits.Total.Value)
	})
}
