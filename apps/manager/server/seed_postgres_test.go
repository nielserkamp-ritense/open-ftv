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

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
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

// Test_Seed_PerFile_AcrossRestarts proves per-file seeding against a real store, across a
// restart: a seed file the operator deleted is not brought back, a seed file the operator
// never touched keeps its content, and a seed file that a later release adds does reach the
// already seeded store.
//
// The rule is language-agnostic (every extension in seedLanguageByExt is seeded the same
// way); the files are cedar only because the test config runs the manager with cedar as its
// PAP language.
func Test_Seed_PerFile_AcrossRestarts(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)
	lgr := slog.New(slog.NewTextHandler(os.Stderr, nil))

	dir := t.TempDir()
	cnf.PAP.Store = dir

	// The policy routes decide on the caller's role, whatever the store holds; the test reads
	// and deletes as an author.
	ip := newIDP(t)
	cnf.OIDC = config2.OIDC{JWKSURL: ip.srv.URL, Issuer: "https://idp.test", Audience: "openftv"}
	author := ip.token(t, "author-user", "author")

	const (
		permitAll = "permit (\n    principal,\n    action,\n    resource\n);\n"
		forbidAll = "forbid (\n    principal,\n    action,\n    resource\n);\n"
	)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "deleted.cedar"), []byte(permitAll), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "kept.cedar"), []byte(permitAll), 0o600))

	deletedID := uuid.NewSHA1(seedNamespace, []byte("deleted.cedar")).String()
	keptID := uuid.NewSHA1(seedNamespace, []byte("kept.cedar")).String()
	addedID := uuid.NewSHA1(seedNamespace, []byte("added.cedar")).String()

	// release 1 seeds deleted.cedar and kept.cedar; the operator then deletes the first on purpose.
	app1 := NewExternal(cnf, lgr).GetMainService().GetFiberApp()

	res := call(t, app1, http.MethodGet, "/v1/policy/"+deletedID, author, nil)
	require.NoError(t, res.Body.Close())
	require.Equal(t, http.StatusOK, res.StatusCode, "deleted.cedar must be seeded on an empty store")

	res = call(t, app1, http.MethodGet, "/v1/policy/"+keptID, author, nil)
	require.NoError(t, res.Body.Close())
	require.Equal(t, http.StatusOK, res.StatusCode, "kept.cedar must be seeded on an empty store")

	res = call(t, app1, http.MethodDelete, "/v1/policy/"+deletedID, author, nil)
	require.NoError(t, res.Body.Close())
	require.Equal(t, http.StatusOK, res.StatusCode)

	// release 2 ships added.cedar as well and changes kept.cedar on disk; the manager restarts
	// against the same store.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "added.cedar"), []byte(permitAll), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "kept.cedar"), []byte(forbidAll), 0o600))

	app2 := NewExternal(cnf, lgr).GetMainService().GetFiberApp()

	res = call(t, app2, http.MethodGet, "/v1/policy/"+deletedID, author, nil)
	require.NoError(t, res.Body.Close())
	require.Equal(t, http.StatusNotFound, res.StatusCode, "a seed file the operator deleted must not come back")

	res = call(t, app2, http.MethodGet, "/v1/policy/"+addedID, author, nil)
	require.NoError(t, res.Body.Close())
	require.Equal(t, http.StatusOK, res.StatusCode, "a seed file added by a later release must reach the populated store")

	res = call(t, app2, http.MethodGet, "/v1/policy/"+keptID, author, nil)
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode, "a seed file the operator left alone must still be there")
	require.Equal(t, permitAll, decodePolicy(t, res).Data, "a present row is the source of truth; a changed seed file must not overwrite it")
}
