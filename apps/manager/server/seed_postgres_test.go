//go:build integration

package server

import (
	"log/slog"
	"os"
	"strings"
	"testing"

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
