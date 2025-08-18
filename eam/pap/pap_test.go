package pap

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("new PAP", func(t *testing.T) {
		h := slog2.NewDummyHandler(0)

		p := New(nil, slog.New(h))
		require.NotNil(t, p)

		assert.NotNil(t, p.logger)
		assert.NotNil(t, p.eventSinks)
		assert.NotNil(t, p.updates)
		assert.NotNil(t, p.deletes)
	})
}
