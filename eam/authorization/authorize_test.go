package authorization

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// upsertStub records or refuses principals.
type upsertStub struct {
	fail error
	seen []identity.Principal
}

func (u *upsertStub) Upsert(_ context.Context, p *identity.Principal) error {
	if u.fail != nil {
		return u.fail
	}

	u.seen = append(u.seen, *p)

	return nil
}

func alice() *RequestPrincipal {
	return &RequestPrincipal{
		Principal: identity.Principal{Kind: identity.KindUser, ID: "alice", Name: "Alice", Email: "alice@example.test", Issuer: "https://idp.test"},
		Roles:     []string{"author"},
	}
}

func TestAuthorize_noAuthPermitsAndRecords(t *testing.T) {
	t.Parallel()

	rec := &upsertStub{}
	a := New(WithLogger(slog.New(slog2.NewDummyHandler(slog.LevelDebug))), NoAuth(), WithPrincipalRecorder(rec)).(*auth)
	resource := models.NewEntity(models.EntityTypePolicy, "p1", models.NewAttributeSet())

	require.NoError(t, a.Authorize(context.Background(), alice(), models.ActionUpdate, resource))
	require.Len(t, rec.seen, 1)
	assert.Equal(t, "alice", rec.seen[0].ID)
}

func TestAuthorize_withoutPermissionsFailsClosed(t *testing.T) {
	t.Parallel()

	rec := &upsertStub{}
	a := New(WithLogger(slog.New(slog2.NewDummyHandler(slog.LevelDebug))), WithPrincipalRecorder(rec)).(*auth)
	resource := models.NewEntity(models.EntityTypePolicy, "p1", models.NewAttributeSet())

	err := a.Authorize(context.Background(), alice(), models.ActionRead, resource)
	require.ErrorIs(t, err, ErrForbidden)
	assert.Empty(t, rec.seen, "a denied caller leaves no personal data behind")
}

func TestAuthorize_recordFailureIsFatalOnlyForWrites(t *testing.T) {
	t.Parallel()

	rec := &upsertStub{fail: errors.New("database is read-only")}
	a := New(WithLogger(slog.New(slog2.NewDummyHandler(slog.LevelDebug))), NoAuth(), WithPrincipalRecorder(rec)).(*auth)
	resource := models.NewEntity(models.EntityTypePolicy, "p1", models.NewAttributeSet())

	require.NoError(t, a.Authorize(context.Background(), alice(), models.ActionRead, resource), "a read must not depend on the principal row")

	err := a.Authorize(context.Background(), alice(), models.ActionUpdate, resource)

	var notRecorded *ErrPrincipalNotRecorded
	require.ErrorAs(t, err, &notRecorded)
}

func TestAuthorize_requiresAResource(t *testing.T) {
	t.Parallel()

	a := New(WithLogger(slog.New(slog2.NewDummyHandler(slog.LevelDebug))), NoAuth()).(*auth)
	require.Error(t, a.Authorize(context.Background(), alice(), models.ActionRead, nil))
}

// A denial is logged with the caller's roles and the action, at the default level, so an
// operator can explain a 403 without redeploying at Debug.
func TestAuthorize_deniedLogsTheRolesAndAction(t *testing.T) {
	t.Parallel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	a := New(WithLogger(slog.New(h)), WithPermissions(ManagerPermissions())).(*auth)
	resource := models.NewEntity(models.EntityTypePolicy, "p1", models.NewAttributeSet())

	auditor := &RequestPrincipal{Principal: identity.NewPrincipal(identity.KindUser, "bob"), Roles: []string{"auditor"}}
	err := a.Authorize(context.Background(), auditor, models.ActionUpdate, resource)
	require.ErrorIs(t, err, ErrForbidden)

	log := h.Log()
	assert.Contains(t, log, "WARN: authorization denied")
	assert.Contains(t, log, "roles=[auditor]")
	assert.Contains(t, log, "action=update")
}

func TestAuthorize_permitsByRoleAndRecords(t *testing.T) {
	t.Parallel()

	rec := &upsertStub{}
	a := New(WithLogger(slog.New(slog2.NewDummyHandler(slog.LevelDebug))), WithPermissions(ManagerPermissions()), WithPrincipalRecorder(rec)).(*auth)
	resource := models.NewEntity(models.EntityTypePolicy, "p1", models.NewAttributeSet())

	require.NoError(t, a.Authorize(context.Background(), alice(), models.ActionUpdate, resource))
	require.Len(t, rec.seen, 1)
	assert.Equal(t, "alice", rec.seen[0].ID)

	require.ErrorIs(t, a.Authorize(context.Background(), nil, models.ActionRead, resource), ErrForbidden, "no caller is the system, which has no role")
}

func TestPermissions_Allows(t *testing.T) {
	t.Parallel()

	user := func(roles ...string) *RequestPrincipal {
		return &RequestPrincipal{Principal: identity.NewPrincipal(identity.KindUser, "u"), Roles: roles}
	}
	app := &RequestPrincipal{Principal: identity.NewPrincipal(identity.KindApp, "key-1")}

	cases := []struct {
		name   string
		p      *Permissions
		caller *RequestPrincipal
		action string
		want   bool
	}{
		{"nil table denies", nil, user("author"), models.ActionRead, false},
		{"nil caller denies", ManagerPermissions(), nil, models.ActionRead, false},
		{"any matching role allows", ManagerPermissions(), user("auditor", "author"), models.ActionDelete, true},
		{"no matching role denies", ManagerPermissions(), user("auditor", "admin"), models.ActionDelete, false},
		{"no roles denies", ManagerPermissions(), user(), models.ActionRead, false},
		{"unknown role denies", ManagerPermissions(), user("visitor"), models.ActionRead, false},
		{"system denies", ManagerPermissions(), SystemRequestPrincipal(), models.ActionRead, false},
		{"unknown principal denies", ManagerPermissions(), &RequestPrincipal{Principal: identity.NewUnknownPrincipal()}, models.ActionRead, false},
		{"app reads", ManagerPermissions(), app, models.ActionRead, true},
		{"app must NOT update", ManagerPermissions(), app, models.ActionUpdate, false},
		{"pap admin does everything", PAPPermissions(), user("admin"), models.ActionDelete, true},
		{"pap user reads", PAPPermissions(), user("user"), models.ActionRead, true},
		{"pap user must NOT update", PAPPermissions(), user("user"), models.ActionUpdate, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.p.Allows(tc.caller, tc.action))
		})
	}
}
