package pap

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("new PAP", func(t *testing.T) {
		h := slog2.NewDummyHandler(0)

		p, err := New(t.Context(), slog.New(h), WithKeyValueDB(memory.New(), ""))
		require.NoError(t, err)
		require.NotNil(t, p)

		assert.NotNil(t, p.logger)
		assert.NotNil(t, p.eventSinks)
		assert.NotNil(t, p.policyUpdates)
		assert.NotNil(t, p.policyDeletes)
	})
}
