package pap

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNewSelfAuthz(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))

	ap, err := NewSelfAuthz(t.Context(), logger, "cedar", "../../testdata/policies", false)
	require.NoError(t, err)
	require.NotNil(t, ap)
}
