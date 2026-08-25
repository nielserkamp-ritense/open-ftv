package management

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// recorder is a decider that records the question and answers as told.
type recorder struct {
	deny     bool
	action   string
	resource *models.Entity
	calls    int
}

func (r *recorder) Decide(_ context.Context, _ *authorization.RequestPrincipal, action string, resource *models.Entity) error {
	r.calls++
	r.action, r.resource = action, resource

	if r.deny {
		return authorization.ErrForbidden
	}

	return nil
}

func newPolicyService(t *testing.T, d *recorder) (*PolicyService, *pap.PAP) {
	t.Helper()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
	store := pap.New(context.Background(), logger, pap.WithLanguage("cedar"))

	return NewPolicyService(logger, store, d), store
}

func author() *authorization.RequestPrincipal {
	return &authorization.RequestPrincipal{Principal: identity.NewPrincipal(identity.KindUser, "alice"), Roles: []string{"author"}}
}

func policy(t *testing.T, id string, tags ...string) *models.Policy {
	t.Helper()

	p, err := models.NewPolicyFromData(id, "cedar", "", "", strings.NewReader("permit(principal, action, resource);"))
	require.NoError(t, err)

	return p.WithTitle(id).WithTags(tags...)
}

func TestPolicyService_List_decidesReadOnTheCollection(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newPolicyService(t, d)

	_, err := svc.List(context.Background(), author())
	require.NoError(t, err)

	assert.Equal(t, 1, d.calls, "one operation is one decision")
	assert.Equal(t, models.ActionRead, d.action)
	assert.Equal(t, "policy::*", d.resource.UID())
}

func TestPolicyService_Get_deniedBeforeNotFound(t *testing.T) {
	t.Parallel()

	d := &recorder{deny: true}
	svc, _ := newPolicyService(t, d)

	_, err := svc.Get(context.Background(), author(), "missing")
	require.ErrorIs(t, err, authorization.ErrForbidden, "a denied caller must not learn whether the id exists")

	assert.Equal(t, "policy::missing", d.resource.UID())
	assert.Nil(t, d.resource.Attributes().GetAttribute(models.AttrStatus), "a missing object contributes its id only")
}

func TestPolicyService_Get_notFoundWhenPermitted(t *testing.T) {
	t.Parallel()

	svc, _ := newPolicyService(t, &recorder{})

	_, err := svc.Get(context.Background(), author(), "missing")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestPolicyService_Create_decidesCreateWithTheIncomingProperties(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, store := newPolicyService(t, d)

	out, err := svc.Create(context.Background(), author(), policy(t, "p1", "laadpalen"), false)
	require.NoError(t, err)
	require.NotNil(t, out)

	assert.Equal(t, models.ActionCreate, d.action)
	assert.Equal(t, "policy::p1", d.resource.UID())
	assert.Equal(t, "concept", d.resource.Attributes().GetAttributeValue(models.AttrStatus))
	assert.Equal(t, []string{"laadpalen"}, d.resource.Attributes().GetAttributeValue(models.AttrTags))
	assert.Equal(t, "cedar", d.resource.Attributes().GetAttributeValue(models.AttrLanguage))

	stored, _, err := store.Read("p1")
	require.NoError(t, err)
	require.NotNil(t, stored)
}

func TestPolicyService_Create_existingIsConflictAfterTheDecision(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newPolicyService(t, d)

	_, err := svc.Create(context.Background(), author(), policy(t, "p2"), false)
	require.NoError(t, err)

	_, err = svc.Create(context.Background(), author(), policy(t, "p2"), false)
	require.ErrorIs(t, err, ErrExists)
	assert.Equal(t, 2, d.calls)

	// forceUpsert turns the create into an update of the stored object.
	_, err = svc.Create(context.Background(), author(), policy(t, "p2"), true)
	require.NoError(t, err)
	assert.Equal(t, models.ActionUpdate, d.action)
}

func TestPolicyService_Update_missingIsNotFoundAfterTheDecision(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newPolicyService(t, d)

	_, err := svc.Update(context.Background(), author(), policy(t, "p1"), false)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Equal(t, models.ActionUpdate, d.action)

	// forceUpsert turns the update into a create.
	_, err = svc.Update(context.Background(), author(), policy(t, "p1"), true)
	require.NoError(t, err)
	assert.Equal(t, models.ActionCreate, d.action)
}

func TestPolicyService_SetStatus_isAcceptOrDeploy(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newPolicyService(t, d)

	_, err := svc.Create(context.Background(), author(), policy(t, "p1"), false)
	require.NoError(t, err)

	// The decision is what this test pins. The deprecated in-memory store refuses the status
	// write itself (PAP.UpdateStatus mutates the object it then hands over as the compare
	// value); the write path is proven against postgres in the manager's integration tests.
	_, err = svc.SetStatus(context.Background(), author(), "p1", models.StatusAccepted)
	if err != nil {
		require.ErrorContains(t, err, "key modified")
	}

	assert.Equal(t, 2, d.calls)
	assert.Equal(t, models.ActionAccept, d.action)
	assert.Equal(t, "concept", d.resource.Attributes().GetAttributeValue(models.AttrStatus), "decided on the stored status, before the change")
}

// The store only reaches deployed through a deployment, so the mapping is pinned on its own.
func TestStatusAction(t *testing.T) {
	t.Parallel()

	assert.Equal(t, models.ActionAccept, statusAction(models.StatusAccepted))
	assert.Equal(t, models.ActionDeploy, statusAction(models.StatusDeployed))
	assert.Equal(t, models.ActionRevert, statusAction(models.StatusConcept))
}

// TestPolicyService_SetStatus_invalidTransitionIsTyped: the life cycle's refusal must reach
// the handler as a client error, not as a server failure.
func TestPolicyService_SetStatus_invalidTransitionIsTyped(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newPolicyService(t, d)

	_, err := svc.Create(context.Background(), author(), policy(t, "p1"), false)
	require.NoError(t, err)

	// concept -> deployed is only reachable through a deployment.
	_, err = svc.SetStatus(context.Background(), author(), "p1", models.StatusDeployed)
	require.ErrorIs(t, err, ErrInvalidTransition)
	assert.Equal(t, models.ActionDeploy, d.action, "decided first, refused by the life cycle second")
}

func TestPolicyService_Delete_decidesDeleteAndHonoursIgnoreMissing(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newPolicyService(t, d)

	_, err := svc.Delete(context.Background(), author(), "missing", false)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Equal(t, models.ActionDelete, d.action)

	out, err := svc.Delete(context.Background(), author(), "missing", true)
	require.NoError(t, err)
	assert.Equal(t, "missing", out.ID())

	d.deny = true
	_, err = svc.Delete(context.Background(), author(), "missing", true)
	require.ErrorIs(t, err, authorization.ErrForbidden, "ignoreMissing never bypasses the decision")
}

func TestPolicyService_nilDeciderPermits(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
	svc := NewPolicyService(logger, pap.New(context.Background(), logger, pap.WithLanguage("cedar")), nil)

	_, err := svc.List(context.Background(), authorization.SystemRequestPrincipal())
	require.NoError(t, err)
}
