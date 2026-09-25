package pip

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNewSelfAuthz(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))

	ip, err := NewSelfAuthz(t.Context(), logger, "../../testdata/unittest/pip2", false, "../../testdata/unittest/pip2/pull")
	require.NoError(t, err)
	require.NotNil(t, ip)
}
