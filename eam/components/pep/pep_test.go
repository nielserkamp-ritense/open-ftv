package pep

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNew(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		h := util.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p := New(nil, logger)
		require.NotNil(t, p)

		p2, ok := p.(*pep)
		require.True(t, ok)
		require.NotNil(t, p2)

		assert.NotNil(t, p2.ctx)
		assert.Equal(t, logger, p2.logger)
	})
}
