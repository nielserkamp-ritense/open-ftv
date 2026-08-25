package server

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config"
	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
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
	ap := pap.New(ctx, logger, pap.WithLanguage("cedar"))

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

// TestSeedAuthzPolicies_noAuditTrail_seedsOnlyEmptyStore guards the fallback for stores
// without an audit trail: such a store cannot tell "never seeded" from "seeded and deleted",
// so it is seeded only while empty. The per-file rule itself is proven against postgres in
// seed_postgres_test.go.
func TestSeedAuthzPolicies_noAuditTrail_seedsOnlyEmptyStore(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "first.cedar"), []byte("permit(principal, action, resource);"), 0o600))

	ctx := context.Background()
	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
	ap := pap.New(ctx, logger, pap.WithLanguage("cedar"))

	s := &Services{
		ctx:    ctx,
		logger: logger,
		pap:    ap,
		cfg:    &config.Config{PAP: config2.PAP{Language: "CEDAR", Store: dir}},
	}

	s.seedAuthzPolicies()

	list, err := ap.List("")
	require.NoError(t, err)
	require.Len(t, list, 1)

	// a later release ships a second file.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "second.cedar"), []byte("forbid(principal, action, resource);"), 0o600))

	s.seedAuthzPolicies()

	list, err = ap.List("")
	require.NoError(t, err)
	require.Len(t, list, 1, "a populated store without an audit trail must not be seeded again")

	second, _, err := ap.Read(uuid.NewSHA1(seedNamespace, []byte("second.cedar")).String())
	require.NoError(t, err)
	require.Nil(t, second)
}

// TestSeedAuthzPolicies_leavesPresentRowUntouched guards "UI wins": a row that is present
// keeps its content even when the seed file on disk has changed since.
func TestSeedAuthzPolicies_leavesPresentRowUntouched(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := filepath.Join(dir, "rule.cedar")
	require.NoError(t, os.WriteFile(file, []byte("permit(principal, action, resource);"), 0o600))

	ctx := context.Background()
	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
	ap := pap.New(ctx, logger, pap.WithLanguage("cedar"))

	s := &Services{
		ctx:    ctx,
		logger: logger,
		pap:    ap,
		cfg:    &config.Config{PAP: config2.PAP{Language: "CEDAR", Store: dir}},
	}

	s.seedAuthzPolicies()

	// the operator edits the row in the UI; meanwhile a release changes the file on disk.
	id := uuid.NewSHA1(seedNamespace, []byte("rule.cedar")).String()
	edited := "permit(principal, action, resource) when { principal has roles };"

	pol, lastIndex, err := ap.Read(id)
	require.NoError(t, err)
	require.NotNil(t, pol)

	pol2, err := models.NewPolicyFromData(id, "cedar", "", "", strings.NewReader(edited))
	require.NoError(t, err)
	_, err = ap.Update(pol, lastIndex, pol2.WithTitle(pol.Title()), seedUser)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(file, []byte("forbid(principal, action, resource);"), 0o600))

	s.seedAuthzPolicies()

	got, _, err := ap.Read(id)
	require.NoError(t, err)
	require.Equal(t, edited, got.ContentString(), "a present row is the source of truth; seeding must not overwrite it")
}
