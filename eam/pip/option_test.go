package pip

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

func TestWithPersistence(t *testing.T) {
	t.Parallel()

	t.Run("with language", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		s := memory.New()

		p := New(nil, logger, WithPersistence(s, ""))
		require.NotNil(t, p)
		assert.Equal(t, s, p.store)
		assert.NotNil(t, p.attributePersist)
	})
}

func TestWithFileStore(t *testing.T) {
	t.Parallel()

	t.Run("with file store", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		path := "../../testdata/pip"

		p := New(nil, logger, WithFileStore(path, true))
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
		p := New(nil, logger, WithPullConfigs("../../testdata/unittest/pip2/pull/empty.yaml"))
		require.NotNil(t, p)
		assert.NotNil(t, p.pullManager)

		h.Clear()

		// empty but good
		p3 := New(nil, logger, WithPullConfigs("../../testdata/unittest/pip2/pull/not_a_valid_file.xyz"))
		require.NotNil(t, p3)
		assert.GreaterOrEqual(t, h.Count(), 2)
		assert.Nil(t, p3.pullManager)
	})
}

func TestWithFactories(t *testing.T) {
	t.Parallel()

	t.Run("with factories", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		b1, b2 := dummyAttributes, dummyEntities

		// with
		p := New(nil, logger, WithFactories(b1, b2))
		require.NotNil(t, p)

		assert.Nil(t, p.NewAttributeSet())
		assert.Nil(t, p.NewEntitySet())

		// without
		p2 := New(nil, logger)
		require.NotNil(t, p2)

		assert.NotNil(t, p2.NewAttributeSet())
		assert.NotNil(t, p2.NewEntitySet())
	})
}

func dummyAttributes(...any) *models.AttributeSet {
	return nil
}

func dummyEntities(...any) *models.EntitySet {
	return nil
}
