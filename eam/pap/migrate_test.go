package pap

import (
	"fmt"
	"log/slog"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestPAP_migration(t *testing.T) {
	t.Parallel()

	t.Run("migration", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		d := "../../testdata/unittest/migrate/sqlite3"

		p, err := New(t.Context(), logger, WithKeyValueDB(memory.New(), ""), WithMigration(d, "sqlite://:memory:", true, 3))
		require.NoError(t, err)
		require.NotNil(t, p)
		h.Clear()

		err = p.Migrate()
		require.NoError(t, err)
		assert.Equalf(t, 2, h.Count(), h.Log())
	})
}

func TestPAP_migration_no_change(t *testing.T) {
	t.Parallel()

	t.Run("migration", func(t *testing.T) {
		tmp := t.TempDir()

		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		d := "../../testdata/unittest/migrate/sqlite3"
		db := fmt.Sprintf("sqlite://%s/test.sqlite3", tmp)

		p, err := New(t.Context(), logger, WithKeyValueDB(memory.New(), ""), WithMigration(d, db, true, 3))
		require.NoError(t, err)
		require.NotNil(t, p)
		h.Clear()

		err = p.Migrate()
		require.NoError(t, err)
		assert.Equalf(t, 2, h.Count(), h.Log())
		h.Clear()

		p2, err := New(t.Context(), logger, WithKeyValueDB(memory.New(), ""), WithMigration(d, db, true, -1))
		require.NoError(t, err)
		require.NotNil(t, p2)
		h.Clear()

		err = p2.Migrate()
		require.NoError(t, err)
		assert.Equalf(t, 1, h.Count(), h.Log())
	})
}

func TestPAP_migration_fail(t *testing.T) {
	t.Parallel()

	t.Run("migration", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		d := "/this/is/not/a/valid/path"

		p, err := New(t.Context(), logger, WithKeyValueDB(memory.New(), ""), WithMigration(d, "sqlite://:memory:", true, 3))
		require.NoError(t, err)
		require.NotNil(t, p)
		h.Clear()

		err = p.Migrate()
		require.Error(t, err)
		assert.Equalf(t, 1, h.Count(), h.Log())
	})
}
