package pap

import (
	"context"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestWithLanguage(t *testing.T) {
	t.Parallel()

	t.Run("with language", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		p, err := New(t.Context(), logger, WithKeyValueDB(memory.New(), ""), WithLanguage("Cedar"))
		require.NoError(t, err)
		require.NotNil(t, p)

		assert.Equal(t, "Cedar", p.language)
		assert.Equal(t, models.CEDAR, p.languageType)
		assert.Equal(t, "Cedar", p.Language().String())
	})
}

func TestWithKeyValueDB(t *testing.T) {
	t.Parallel()

	t.Run("with key/value DB", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		s := memory.New()

		p, err := New(t.Context(), logger, WithKeyValueDB(s, ""))
		require.NoError(t, err)
		require.NotNil(t, p)

		assert.Equal(t, s, p.kvStore)
		assert.NotNil(t, p.policyDB)
		assert.NotNil(t, p.bundleDB)
	})
}

func TestWithPgPool(t *testing.T) {
	t.Parallel()

	t.Run("with postgres pool", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		pool, err := postgresql.New(ctx, "postgres://localhost:5432/myDB", time.Minute, 3)
		require.NoError(t, err)

		p, err := New(t.Context(), logger, WithPgPool(pool))
		require.NoError(t, err)
		require.NotNil(t, p)

		assert.Nil(t, p.kvStore)
		assert.NotNil(t, p.languageDB)
		assert.NotNil(t, p.policyDB)
	})
}

func TestWithPolicyDB(t *testing.T) {
	t.Parallel()

	t.Run("with policy DB", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		db, err := NewPolicyDB(ctx, "postgres://localhost:5432/myDB", time.Minute, 3)
		require.NoError(t, err)

		p, err := New(t.Context(), logger, WithPolicyDB(db))
		require.NoError(t, err)
		require.NotNil(t, p)

		assert.Nil(t, p.kvStore)
		assert.Equal(t, db, p.policyDB)
	})
}

func TestWithFileStore(t *testing.T) {
	t.Parallel()

	t.Run("with file store", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		path := "../../testdata/policies/opa"

		p, err := New(t.Context(), logger, WithKeyValueDB(memory.New(), ""), WithFileStore(path, true))
		require.NoError(t, err)
		require.NotNil(t, p)

		path2, err := filepath.Abs(path)
		require.NoError(t, err)

		assert.Equal(t, path2, p.policyStore)
		assert.True(t, p.recurse)
	})
}

func TestWithMigration(t *testing.T) {
	t.Parallel()

	t.Run("with migration", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		p, err := New(t.Context(), logger, WithKeyValueDB(memory.New(), ""), WithMigration("/file", "/db", true, 3))
		require.NoError(t, err)
		require.NotNil(t, p)

		assert.Equal(t, "/file", p.migrateSource)
		assert.Equal(t, "/db", p.migrateDB)
		assert.True(t, p.migrateAuto)
		assert.Equal(t, 3, p.migrateSteps)
	})
}
