package opensearch

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateIndex(t *testing.T) {
	t.Run("create & delete index", func(t *testing.T) {
		b, err := newBase(user, pswd, endpoints)
		require.NoError(t, err)
		require.NotNil(t, b)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		u, _ := uuid.NewUUID()
		name := strings.Replace(u.String(), "-", "", -1)

		err = b.CreateIndex(ctx, name, 1, 1)
		require.NoError(t, err)

		err = b.DeleteIndexes(ctx, name)
		require.NoError(t, err)
	})
}
