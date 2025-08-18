package migrate

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestPostgres(t *testing.T) {
	t.Parallel()

	t.Run("migrate postgres", func(t *testing.T) {
		tmp := t.TempDir()

		d, err := filepath.Abs("../../testdata/unittest/migrate/sqlite3")
		require.NoError(t, err)

		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		db := fmt.Sprintf("sqlite://%s/test.sqlite3", tmp)

		err = Postgres(fmt.Sprintf("file://%s", d), db, 1, logger)
		require.NoError(t, err)
		assert.Equalf(t, 1, h.Count(), h.Log())
		h.Clear()

		err = Postgres(fmt.Sprintf("file://%s", d), db, -1, logger)
		require.NoError(t, err)
		assert.Equalf(t, 1, h.Count(), h.Log())
		h.Clear()

		err = Postgres(fmt.Sprintf("file://%s", d), db, 0, logger)
		require.NoError(t, err)
		assert.Equalf(t, 1, h.Count(), h.Log())
		h.Clear()

		err = Postgres("file:///this/is/not/a/valid/directory", db, 0, logger)
		require.Error(t, err)
	})
}
