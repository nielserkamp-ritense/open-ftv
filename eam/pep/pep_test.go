package pep

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("new", func(t *testing.T) {
		h := util.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p := New(nil, logger)
		require.NotNil(t, p)

		assert.NotNil(t, p.ctx)
		assert.Equal(t, logger, p.logger)
	})
}
