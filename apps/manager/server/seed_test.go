package server

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config"
	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// TestSeedAuthzPolicies_UniqueTitles guards against the policy_ix1 UNIQUE(language, title)
// collision: every seeded cedar policy must get a distinct, non-empty title, otherwise the
// postgres-backed store keeps only the first and silently drops the rest (e.g. the role
// policies), leaving every role-based request to default-deny.
func TestSeedAuthzPolicies_UniqueTitles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
	ap := pap.New(ctx, logger, pap.WithLanguage("cedar"))

	s := &Services{
		ctx:    ctx,
		logger: logger,
		pap:    ap,
		cfg: &config.Config{PAP: config2.PAP{
			Language: "CEDAR",
			Store:    "../../../testdata/apps/manager/policies/cedar",
		}},
	}

	s.seedAuthzPolicies()

	list, err := ap.List("")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(list), 6, "all bundled cedar policies should be seeded")

	seen := map[string]bool{}
	for _, p := range list {
		title := p.Title()
		require.NotEmpty(t, title, "seeded policy %s has an empty title (would collide on policy_ix1)", p.ID())
		require.False(t, seen[title], "duplicate seeded policy title %q (collides on policy_ix1)", title)
		seen[title] = true
	}
}
