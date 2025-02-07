package opensearch

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger_Log(t *testing.T) {
	t.Run("log", func(t *testing.T) {
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

		err = l.Log(ctx, false, LogRecord{
			Index: index,
			Data:  map[string]any{"hello": "world", "int": 1, "bool": true},
		})
		require.NoError(t, err)
	})
}

func TestLogger_LogBulk(t *testing.T) {
	t.Run("log bulk", func(t *testing.T) {
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

		err = l.LogBulk(ctx, false,
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
	})
}
