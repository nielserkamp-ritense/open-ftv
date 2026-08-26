//go:build integration

package server_test

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/server"
)

// seedNamespace mirrors the unexported namespace in seed.go, so the test can address a
// seeded policy by the id the manager gives it.
var seedNamespace = uuid.NewSHA1(uuid.NameSpaceURL, []byte("openftv-mgmt-authz-policies"))

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
	app1 := server.NewExternal(cnf, lgr).GetMainService().GetFiberApp()

	require.Equal(t, http.StatusOK, getPolicy(t, app1, firstID).StatusCode, "first.cedar must be seeded on an empty store")
	require.Equal(t, http.StatusOK, deletePolicy(t, app1, firstID).StatusCode)

	// release 2 ships second.cedar as well and the manager restarts against the same store.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "second.cedar"), []byte(permitAll), 0o600))

	app2 := server.NewExternal(cnf, lgr).GetMainService().GetFiberApp()

	require.Equal(t, http.StatusNotFound, getPolicy(t, app2, firstID).StatusCode, "a seed file the operator deleted must not come back")
	require.Equal(t, http.StatusOK, getPolicy(t, app2, secondID).StatusCode, "a seed file added by a later release must reach the populated store")
}
