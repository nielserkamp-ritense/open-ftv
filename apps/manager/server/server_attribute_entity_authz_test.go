//go:build integration

package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

// Test_AttributeAndEntity_RoleMatrix drives the SCRUM-16 matrix for PIP data through the
// manager as deployed: secured mode, OIDC, the shipped policies seeded into postgres. The
// author creates, updates, deletes, accepts, reverts and restores; every role reads; admin and
// auditor change nothing; an accepted object is not edited in place.
func Test_AttributeAndEntity_RoleMatrix(t *testing.T) {
	cnf := newManagerConfigWithPostgres(t)
	cnf.PAP.Store = "../../../testdata/apps/manager/policies/cedar"
	cnf.Authorization.FailClosedOnEmpty = true

	ip := newIDP(t)
	cnf.OIDC = config2.OIDC{JWKSURL: ip.srv.URL, Issuer: "https://idp.test", Audience: "openftv"}

	// the log is kept: the detail view only warns when its audit or usage query fails.
	logged := &syncBuffer{}
	lgr := slog.New(slog.NewTextHandler(io.MultiWriter(os.Stderr, logged), nil))
	app := NewExternal(cnf, lgr).GetMainService().GetFiberApp()

	author, auditor, admin := ip.token(t, "author-user", "author"), ip.token(t, "auditor-user", "auditor"), ip.token(t, "admin-user", "admin")
	everyone := map[string]string{"author": author, "auditor": auditor, "admin": admin}

	attr := oas.Attribute{Key: "max-age", Type: "string", Value: "18", Status: "concept", Metadata: oas.Metadata{Title: "Maximum age", Tags: []string{"laadpalen"}}}
	// The entity carries an attribute: the postgres PIP cannot read back an entity stored
	// without any (nil attributes column), which is outside this test.
	ent := oas.Entity{Type: "user", Id: "alice", Status: "concept", Metadata: oas.Metadata{Title: "Alice"},
		Attributes: []oas.Attribute{{Key: "age", Type: "integer", Value: 42}}}

	t.Run("everyone reads attributes and entities", func(t *testing.T) {
		for name, tok := range everyone {
			require.Equal(t, http.StatusOK, call(t, app, http.MethodGet, "/v1/attributes", tok, nil).StatusCode, name)
			require.Equal(t, http.StatusOK, call(t, app, http.MethodGet, "/v1/entities", tok, nil).StatusCode, name)
		}
	})

	t.Run("no token is denied, not served", func(t *testing.T) {
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodGet, "/v1/attributes", "", nil).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodGet, "/v1/entities", "", nil).StatusCode)
	})

	t.Run("only the author creates PIP data", func(t *testing.T) {
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPost, "/v1/attribute/max-age", admin, attr).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPost, "/v1/attribute/max-age", auditor, attr).StatusCode)
		require.Equal(t, http.StatusCreated, call(t, app, http.MethodPost, "/v1/attribute/max-age", author, attr).StatusCode)

		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPost, "/v1/entity/user/alice", admin, ent).StatusCode)
		require.Equal(t, http.StatusCreated, call(t, app, http.MethodPost, "/v1/entity/user/alice", author, ent).StatusCode)
	})

	t.Run("the detail view reads its own audit log and usage data", func(t *testing.T) {
		require.Equal(t, http.StatusOK, call(t, app, http.MethodGet, "/v1/attribute/max-age", auditor, nil).StatusCode)
		require.Equal(t, http.StatusOK, call(t, app, http.MethodGet, "/v1/entity/user/alice", auditor, nil).StatusCode)

		// a failing query is a warning and an empty list, never a failed request: only the log tells.
		require.NotContains(t, logged.String(), "failed to read attribute")
		require.NotContains(t, logged.String(), "failed to read entity")
	})

	t.Run("a user-only principal reads no PIP data", func(t *testing.T) {
		// the manager's examples dropped the user role (ADR 0006): no rule grants it a read.
		user := ip.token(t, "plain-user", "user")
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodGet, "/v1/attributes", user, nil).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodGet, "/v1/entities", user, nil).StatusCode)
	})

	t.Run("a denied caller cannot tell a missing object from an existing one", func(t *testing.T) {
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodDelete, "/v1/attribute/max-age", admin, nil).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodDelete, "/v1/attribute/nope", admin, nil).StatusCode)
		require.Equal(t, http.StatusNotFound, call(t, app, http.MethodDelete, "/v1/attribute/nope", author, nil).StatusCode)
	})

	t.Run("a status change is accept or revert, never a read", func(t *testing.T) {
		accept := oas.AttributeStatus{Key: "max-age", Status: "accepted"}
		revert := oas.AttributeStatus{Key: "max-age", Status: "concept"}

		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPatch, "/v1/attribute/max-age/status", auditor, accept).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPatch, "/v1/attribute/max-age/status", admin, accept).StatusCode)
		require.Equal(t, http.StatusOK, call(t, app, http.MethodPatch, "/v1/attribute/max-age/status", author, accept).StatusCode)

		// accepted is not edited in place; revert it first.
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPut, "/v1/attribute/max-age", author, attr).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPatch, "/v1/attribute/max-age/status", auditor, revert).StatusCode)
		require.Equal(t, http.StatusOK, call(t, app, http.MethodPatch, "/v1/attribute/max-age/status", author, revert).StatusCode)
		require.Equal(t, http.StatusOK, call(t, app, http.MethodPut, "/v1/attribute/max-age", author, attr).StatusCode)

		// the same life cycle for an entity.
		eAccept := oas.EntityStatus{Type: "user", Id: "alice", Status: "accepted"}
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPatch, "/v1/entity/user/alice/status", admin, eAccept).StatusCode)
		require.Equal(t, http.StatusOK, call(t, app, http.MethodPatch, "/v1/entity/user/alice/status", author, eAccept).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPut, "/v1/entity/user/alice", author, ent).StatusCode)

		// an invalid transition is the caller's error, once permitted; a body without a status,
		// or with one PIP data cannot take, never reaches a decision.
		require.Equal(t, http.StatusBadRequest, call(t, app, http.MethodPatch, "/v1/entity/user/alice/status", author, eAccept).StatusCode)
		require.Equal(t, http.StatusBadRequest, call(t, app, http.MethodPatch, "/v1/attribute/max-age/status", auditor, oas.AttributeStatus{Key: "max-age"}).StatusCode)
		require.Equal(t, http.StatusBadRequest, call(t, app, http.MethodPatch, "/v1/attribute/max-age/status", auditor, oas.AttributeStatus{Key: "max-age", Status: "deployed"}).StatusCode)
	})

	t.Run("only the author restores a version", func(t *testing.T) {
		// Versions are written by a deployment, so none exists here: the denied roles get 403
		// before the 404 the author gets, which is the decision this pins.
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPost, "/v1/attribute/max-age/version/1/restore", auditor, nil).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPost, "/v1/attribute/max-age/version/1/restore", admin, nil).StatusCode)
		require.Equal(t, http.StatusNotFound, call(t, app, http.MethodPost, "/v1/attribute/max-age/version/1/restore", author, nil).StatusCode)
		// The author's entity restore is not asserted: the postgres PIP reads entity versions
		// from the attribute_version table (SQLSTATE 42703), which predates this test.
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodPost, "/v1/entity/user/alice/version/1/restore", admin, nil).StatusCode)
	})

	t.Run("an edit is stored as a concept whatever status the body names", func(t *testing.T) {
		// Reaching deployed needs a bundle target, which this server has none of; the guard's
		// answer for deployed data is pinned in eam/management's matrix. What the API adds is
		// that no write smuggles a status in: only accept makes PIP data accepted.
		statusOf := func(path string) string {
			res := call(t, app, http.MethodGet, path, auditor, nil)
			require.Equal(t, http.StatusOK, res.StatusCode)

			var out struct {
				Status string `json:"status"`
			}
			require.NoError(t, json.NewDecoder(res.Body).Decode(&out))

			return out.Status
		}

		revert := oas.EntityStatus{Type: "user", Id: "alice", Status: "concept"}
		require.Equal(t, http.StatusOK, call(t, app, http.MethodPatch, "/v1/entity/user/alice/status", author, revert).StatusCode)

		body := ent
		body.Status = "deployed"
		require.Equal(t, http.StatusOK, call(t, app, http.MethodPut, "/v1/entity/user/alice", author, body).StatusCode)
		require.Equal(t, "concept", statusOf("/v1/entity/user/alice"))

		born := ent
		born.Id, born.Status = "bob", "accepted"
		require.Equal(t, http.StatusCreated, call(t, app, http.MethodPost, "/v1/entity/user/bob", author, born).StatusCode)
		require.Equal(t, "concept", statusOf("/v1/entity/user/bob"))
	})

	t.Run("only the author deletes", func(t *testing.T) {
		// The author's delete itself is not asserted: on postgres it fails on the foreign key
		// from attribute_audit and entity_audit to their object (the policy_audit key was
		// dropped in migration 13, these were not), which predates this test.
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodDelete, "/v1/entity/user/alice", auditor, nil).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodDelete, "/v1/entity/user/alice", admin, nil).StatusCode)
		require.Equal(t, http.StatusForbidden, call(t, app, http.MethodDelete, "/v1/attribute/max-age", auditor, nil).StatusCode)
	})
}

// syncBuffer is a bytes.Buffer the server's goroutines may write to while the test reads it.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}
