package pap

import (
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/storage/valkeyrie/memory"
)

func TestWithLanguage(t *testing.T) {
	t.Run("with language", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		p := New(nil, logger, WithLanguage("Cedar"))
		require.NotNil(t, p)

		p2, ok := p.(*pap)
		require.True(t, ok)
		require.NotNil(t, p2)
		assert.Equal(t, "Cedar", p2.language)
	})
}

func TestWithPersistence(t *testing.T) {
	t.Run("with language", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		s := memory.New()

		p := New(nil, logger, WithPersistence(s, ""))
		require.NotNil(t, p)

		p2, ok := p.(*pap)
		require.True(t, ok)
		require.NotNil(t, p2)
		assert.Equal(t, s, p2.store)
		assert.NotNil(t, p2.persist)
	})
}

func TestWithFileStore(t *testing.T) {
	t.Run("with file store", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		path := "../../../testdata/policies/opa"

		p := New(nil, logger, WithFileStore(path, true))
		require.NotNil(t, p)

		path2, err := filepath.Abs(path)
		require.NoError(t, err)

		p2, ok := p.(*pap)
		require.True(t, ok)
		require.NotNil(t, p2)
		assert.Equal(t, path2, p2.policyStore)
		assert.True(t, p2.recurse)
	})
}
