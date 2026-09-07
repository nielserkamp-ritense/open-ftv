package pip

import (
	"context"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestWithKeyValueDB(t *testing.T) {
	t.Parallel()

	t.Run("with key/value db", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		s := memory.New()

		p, err := New(t.Context(), logger, WithKeyValueDB(s, ""))
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.Equal(t, s, p.kvStore)
		assert.NotNil(t, p.attributeDB)
		assert.NotNil(t, p.entityDB)
	})
}

func TestWithPostgresDB(t *testing.T) {
	t.Parallel()

	t.Run("with postgres db", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		db, err := postgresql.New(ctx, "postgres://localhost:5432/open_ftv", time.Microsecond, 2)
		require.NoError(t, err)
		require.NotNil(t, db)

		p, err := New(t.Context(), logger, WithPostgresDB(NewPostgresWithPool(db)))
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.Nil(t, p.kvStore)
		assert.NotNil(t, p.attributeDB)
		assert.NotNil(t, p.entityDB)
	})
}

func TestWithFileStore(t *testing.T) {
	t.Parallel()

	t.Run("with file store", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		path := "../../testdata/pip"

		p, err := New(t.Context(), logger, WithKeyValueDB(memory.New(), ""), WithFileStore(path, true))
		require.NoError(t, err)
		require.NotNil(t, p)

		path2, err := filepath.Abs(path)
		require.NoError(t, err)

		assert.Equal(t, path2+"/attributes", p.attrStore)
		assert.Equal(t, path2+"/entities", p.entityStore)
		assert.True(t, p.recurse)
	})
}

func TestWithPullConfigs(t *testing.T) {
	t.Parallel()

	t.Run("with pull configs", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		// empty but good
		p, err := New(t.Context(), logger, WithKeyValueDB(memory.New(), ""), WithPullConfigs("../../testdata/unittest/pip2/pull/empty.yaml"))
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.NotNil(t, p.pullManager)

		h.Clear()

		// empty but good
		p3, err := New(t.Context(), logger, WithKeyValueDB(memory.New(), ""), WithPullConfigs("../../testdata/unittest/pip2/pull/not_a_valid_file.xyz"))
		require.NoError(t, err)
		require.NotNil(t, p3)
		assert.GreaterOrEqual(t, h.Count(), 2)
		assert.Nil(t, p3.pullManager)
	})
}
