package pap

import (
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestWithLanguage(t *testing.T) {
	t.Parallel()

	t.Run("with language", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		p := New(nil, logger, WithLanguage("Cedar"))
		require.NotNil(t, p)

		assert.Equal(t, "Cedar", p.language)
		assert.Equal(t, models.CEDAR, p.languageType)
		assert.Equal(t, "Cedar", p.Language().String())
	})
}

func TestWithPersistence(t *testing.T) {
	t.Parallel()

	t.Run("with language", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		s := memory.New()

		p := New(nil, logger, WithPersistence(s, ""))
		require.NotNil(t, p)

		assert.Equal(t, s, p.store)
		assert.NotNil(t, p.persist)
	})
}

func TestWithFileStore(t *testing.T) {
	t.Parallel()

	t.Run("with file store", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		path := "../../testdata/policies/opa"

		p := New(nil, logger, WithFileStore(path, true))
		require.NotNil(t, p)

		path2, err := filepath.Abs(path)
		require.NoError(t, err)

		assert.Equal(t, path2, p.policyStore)
		assert.True(t, p.recurse)
	})
}
