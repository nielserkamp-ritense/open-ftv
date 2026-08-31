package server

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config"
	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestSeedAuthzPolicies_readsMetaTags(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "generic.cedar"), []byte("permit(principal, action, resource);"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "generic.cedar.meta"), []byte(`metadata:
  tags:
    - rvig
    - rdw
`), 0o600))

	ctx := context.Background()
	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
	ap, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"))
	require.NoError(t, err)

	s := &Services{
		ctx:    ctx,
		logger: logger,
		pap:    ap,
		cfg: &config.Config{PAP: config2.PAP{
			Language: "CEDAR",
			Store:    dir,
		}},
	}

	s.seedAuthzPolicies()

	id := uuid.NewSHA1(seedNamespace, []byte("generic.cedar")).String()
	got, _, err := ap.Read(id)
	require.NoError(t, err)
	require.True(t, got.HasTag("rvig"))
	require.True(t, got.HasTag("rdw"))
}
