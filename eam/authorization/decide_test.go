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

func TestManagementPARC_carriesOnlyDecisionInput(t *testing.T) {
	t.Parallel()

	ctx := WithTrace(context.Background(), Trace{Parent: "00-abc-def-01", State: "vendor=1"})
	resource := models.NewEntity(models.EntityTypePolicy, "p1", models.NewAttributeSet())

	parc := ManagementPARC(ctx, alice(), models.ActionUpdate, resource)

	assert.Equal(t, "user::alice", parc.Principal.UID())
	assert.Equal(t, []string{"author"}, parc.Principal.Attributes().GetAttributeValue(models.AttrRoles))
	assert.Nil(t, parc.Principal.Attributes().GetAttribute(models.AttrEmail), "display claims never reach the PDP")
	assert.Nil(t, parc.Principal.Attributes().GetAttribute(models.AttrPreferredName))
	assert.Nil(t, parc.Principal.Attributes().GetAttribute(models.AttrIssuer))

	assert.Equal(t, "Action::update", parc.Action.UID())
	assert.Same(t, resource, parc.Resource)

	assert.NotNil(t, parc.Context.GetAttribute(models.AttrTime))
	assert.Equal(t, "00-abc-def-01", parc.Context.GetAttributeValue(models.AttrTraceParent))
	assert.Equal(t, "vendor=1", parc.Context.GetAttributeValue(models.AttrTraceState))
	assert.Nil(t, parc.Context.GetAttribute(models.AttrHTTP), "no request shape in the context")
	assert.Nil(t, parc.Context.GetAttribute(models.AttrJWT))
}

func TestDecide_noAuthPermitsAndRecords(t *testing.T) {
	t.Parallel()

	rec := &upsertStub{}
	a := New(WithLogger(slog.New(slog2.NewDummyHandler(slog.LevelDebug))), NoAuth(), WithPrincipalRecorder(rec)).(*auth)
	resource := models.NewEntity(models.EntityTypePolicy, "p1", models.NewAttributeSet())

	require.NoError(t, a.Decide(context.Background(), alice(), models.ActionUpdate, resource))
	require.Len(t, rec.seen, 1)
	assert.Equal(t, "alice", rec.seen[0].ID)
}

func TestDecide_withoutPDPFailsClosed(t *testing.T) {
	t.Parallel()

	rec := &upsertStub{}
	a := New(WithLogger(slog.New(slog2.NewDummyHandler(slog.LevelDebug))), WithPrincipalRecorder(rec)).(*auth)
	resource := models.NewEntity(models.EntityTypePolicy, "p1", models.NewAttributeSet())

	err := a.Decide(context.Background(), alice(), models.ActionRead, resource)
	require.ErrorIs(t, err, ErrForbidden)
	assert.Empty(t, rec.seen, "a denied caller leaves no personal data behind")
}

func TestDecide_recordFailureIsFatalOnlyForWrites(t *testing.T) {
	t.Parallel()

	rec := &upsertStub{fail: errors.New("database is read-only")}
	a := New(WithLogger(slog.New(slog2.NewDummyHandler(slog.LevelDebug))), NoAuth(), WithPrincipalRecorder(rec)).(*auth)
	resource := models.NewEntity(models.EntityTypePolicy, "p1", models.NewAttributeSet())

	require.NoError(t, a.Decide(context.Background(), alice(), models.ActionRead, resource), "a read must not depend on the principal row")

	err := a.Decide(context.Background(), alice(), models.ActionUpdate, resource)

	var notRecorded *ErrPrincipalNotRecorded
	require.ErrorAs(t, err, &notRecorded)
}

func TestDecide_requiresAResource(t *testing.T) {
	t.Parallel()

	a := New(WithLogger(slog.New(slog2.NewDummyHandler(slog.LevelDebug))), NoAuth()).(*auth)
	require.Error(t, a.Decide(context.Background(), alice(), models.ActionRead, nil))
}
