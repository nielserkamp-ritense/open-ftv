//go:build integration

package server

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// TestSeedAuthzPolicies_UniqueTitles guards against the policy_ix1 UNIQUE(language, title)
// collision: a duplicate title means postgres keeps only the first seeded policy and drops
// the rest, leaving role-based requests to default-deny.
//
// Needs a real postgres backend: the in-memory KV store keys policies by id, not
// (language, title), so it can't reproduce this collision.
func TestSeedAuthzPolicies_UniqueTitles(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)
	cnf.PAP.Store = "../../../testdata/apps/manager/policies/cedar"

	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))

	srv := NewExternal(cnf, lgr)

	list, err := srv.pap.List("")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(list), 6, "all bundled cedar policies should be seeded")

	seen := map[string]bool{}

	for _, p := range list {
		title := p.Title()
		require.NotEmpty(t, title, "seeded policy %s has an empty title (would collide on policy_ix1)", p.ID())
		require.False(t, seen[title], "duplicate seeded policy title %q (collides on policy_ix1)", title)
		seen[title] = true
	}

	// Prove policy_ix1 actually exists and is enforced: a second policy that reuses an
	// already-seeded (language, title) pair must be rejected by postgres, not silently stored.
	existing := list[0]

	dup, err := models.NewPolicyFromData(
		"11111111-1111-1111-1111-111111111111",
		existing.Language(), "", "",
		strings.NewReader("permit(principal, action, resource);"),
	)
	require.NoError(t, err)

	dup = dup.WithTitle(existing.Title())

	_, err = srv.pap.Create(dup, seedUser)
	require.Error(t, err, "policy_ix1 should reject a duplicate (language, title) pair")
}

// Test_Seed_PerFile_AcrossRestarts proves the two halves of per-file seeding against a real
// store: a seed file the operator deleted is not brought back by a restart, and a seed file
// that a later release adds does reach the already seeded store.
func Test_Seed_PerFile_AcrossRestarts(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)
	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))

	dir := t.TempDir()
	cnf.PAP.Store = dir

	const permitAll = "permit (\n    principal,\n    action,\n    resource\n);\n"

	require.NoError(t, os.WriteFile(filepath.Join(dir, "first.cedar"), []byte(permitAll), 0o600))

	firstID := uuid.NewSHA1(seedNamespace, []byte("first.cedar")).String()
	secondID := uuid.NewSHA1(seedNamespace, []byte("second.cedar")).String()

	// release 1 seeds first.cedar; the operator then deletes it on purpose.
	app1 := NewExternal(cnf, lgr).GetMainService().GetFiberApp()

	require.Equal(t, http.StatusOK, getPolicy(t, app1, firstID).StatusCode, "first.cedar must be seeded on an empty store")
	require.Equal(t, http.StatusOK, deletePolicy(t, app1, firstID).StatusCode)

	// release 2 ships second.cedar as well and the manager restarts against the same store.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "second.cedar"), []byte(permitAll), 0o600))

	app2 := NewExternal(cnf, lgr).GetMainService().GetFiberApp()

	require.Equal(t, http.StatusNotFound, getPolicy(t, app2, firstID).StatusCode, "a seed file the operator deleted must not come back")
	require.Equal(t, http.StatusOK, getPolicy(t, app2, secondID).StatusCode, "a seed file added by a later release must reach the populated store")
}
