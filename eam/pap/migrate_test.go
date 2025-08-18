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
)

func TestPAP_migration(t *testing.T) {
	t.Parallel()

	t.Run("migration", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		d := "../../testdata/unittest/migrate/sqlite3"

		// we don't supply the persistence option, so the automatic migration is skipped.
		p := New(nil, logger, WithMigration(d, "sqlite://:memory:", true, 3))
		require.NotNil(t, p)
		h.Clear()

		// now we can call it separately.
		err := p.migration()
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

		// we don't supply the persistence option, so the automatic migration is skipped.
		p := New(nil, logger, WithMigration(d, db, true, 3))
		require.NotNil(t, p)
		h.Clear()

		// now we can call it separately.
		err := p.migration()
		require.NoError(t, err)
		assert.Equalf(t, 2, h.Count(), h.Log())
		h.Clear()

		// we don't supply the persistence option, so the automatic migration is skipped.
		p2 := New(nil, logger, WithMigration(d, db, true, -1))
		require.NotNil(t, p2)
		h.Clear()

		// now we can call it separately.
		err = p2.migration()
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

		// we don't supply the persistence option, so the automatic migration is skipped.
		p := New(nil, logger, WithMigration(d, "sqlite://:memory:", true, 3))
		require.NotNil(t, p)
		h.Clear()

		// now we can call it separately.
		err := p.migration()
		require.Error(t, err)
		assert.Equalf(t, 1, h.Count(), h.Log())
	})
}
