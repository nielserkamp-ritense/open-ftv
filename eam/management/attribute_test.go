package management

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func newAttributeService(d *recorder) (*AttributeService, *memAttributes) {
	store := newMemAttributes()
	return NewAttributeService(slog.New(slog2.NewDummyHandler(slog.LevelDebug)), store, d), store
}

// attribute is the submitted form of the attribute every test works on.
func attribute(tags ...string) *oas.Attribute {
	return &oas.Attribute{Key: "max-age", Type: "string", Value: "v", Metadata: oas.Metadata{Title: "max-age", Tags: tags}}
}

func TestAttributeService_List_decidesReadOnTheCollection(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newAttributeService(d)

	_, err := svc.List(context.Background(), author())
	require.NoError(t, err)

	assert.Equal(t, 1, d.calls, "one operation is one decision")
	assert.Equal(t, models.ActionRead, d.action)
	assert.Equal(t, "attribute::*", d.resource.UID())
}

func TestAttributeService_Get_deniedBeforeNotFound(t *testing.T) {
	t.Parallel()

	d := &recorder{deny: true}
	svc, _ := newAttributeService(d)

	_, err := svc.Get(context.Background(), author(), "missing")
	require.ErrorIs(t, err, authorization.ErrForbidden, "a denied caller must not learn whether the key exists")

	assert.Equal(t, 1, d.calls)
	assert.Equal(t, "attribute::missing", d.resource.UID())
	assert.Nil(t, d.resource.Attributes().GetAttribute(models.AttrStatus), "a missing object contributes its key only")
}

func TestAttributeService_Get_notFoundWhenPermitted(t *testing.T) {
	t.Parallel()

	svc, _ := newAttributeService(&recorder{})

	_, err := svc.Get(context.Background(), author(), "missing")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestAttributeService_Create_decidesCreateWithTheSubmittedProperties(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, store := newAttributeService(d)

	out, err := svc.Create(context.Background(), author(), attribute("laadpalen"), false)
	require.NoError(t, err)
	require.NotNil(t, out)

	assert.Equal(t, models.ActionCreate, d.action)
	assert.Equal(t, "attribute::max-age", d.resource.UID())
	assert.Equal(t, "concept", d.resource.Attributes().GetAttribute(models.AttrStatus).Value(), "a new attribute is stored as concept")
	assert.Equal(t, []string{"laadpalen"}, d.resource.Attributes().GetAttribute(models.AttrTags).Value())

	stored, _, err := store.GetAttribute("max-age")
	require.NoError(t, err)
	assert.Equal(t, "alice", stored.CreatedBy(), "the write is attributed to the caller")
}

func TestAttributeService_Create_existingIsConflictAfterTheDecision(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newAttributeService(d)

	_, err := svc.Create(context.Background(), author(), attribute(), false)
	require.NoError(t, err)

	_, err = svc.Create(context.Background(), author(), attribute(), false)
	require.ErrorIs(t, err, ErrExists)
	assert.Equal(t, 2, d.calls, "the conflict is reported after the decision, not instead of it")
	assert.Equal(t, models.ActionCreate, d.action)

	_, err = svc.Create(context.Background(), author(), attribute(), true)
	require.NoError(t, err)
	assert.Equal(t, models.ActionUpdate, d.action, "forceUpsert on an existing key is an update")
}

func TestAttributeService_Update_decidesOnTheStoredProperties(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, store := newAttributeService(d)

	_, err := svc.Create(context.Background(), author(), attribute("laadpalen"), false)
	require.NoError(t, err)

	_, err = store.UpdateAttributeStatus("max-age", models.StatusAccepted, identity.NewSystemPrincipal())
	require.NoError(t, err)

	_, err = svc.Update(context.Background(), author(), attribute("other"), false)
	require.NoError(t, err)

	assert.Equal(t, models.ActionUpdate, d.action)
	assert.Equal(t, "accepted", d.resource.Attributes().GetAttribute(models.AttrStatus).Value(), "the PDP sees the stored status, not the submitted one")
	assert.Equal(t, []string{"laadpalen"}, d.resource.Attributes().GetAttribute(models.AttrTags).Value())
}

func TestAttributeService_Update_missingIsNotFoundAfterTheDecision(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newAttributeService(d)

	_, err := svc.Update(context.Background(), author(), attribute(), false)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Equal(t, 1, d.calls)
	assert.Equal(t, models.ActionUpdate, d.action)

	_, err = svc.Update(context.Background(), author(), attribute(), true)
	require.NoError(t, err)
	assert.Equal(t, models.ActionCreate, d.action, "forceUpsert on a missing key is a create")
}

func TestAttributeService_SetStatus_isAcceptOrRevert(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newAttributeService(d)

	_, err := svc.Create(context.Background(), author(), attribute(), false)
	require.NoError(t, err)

	out, err := svc.SetStatus(context.Background(), author(), "max-age", models.StatusAccepted)
	require.NoError(t, err)
	assert.Equal(t, models.ActionAccept, d.action)
	assert.Equal(t, models.StatusAccepted, out.Status())

	out, err = svc.SetStatus(context.Background(), author(), "max-age", models.StatusConcept)
	require.NoError(t, err)
	assert.Equal(t, models.ActionRevert, d.action)
	assert.Equal(t, models.StatusConcept, out.Status())
}

func TestAttributeService_SetStatus_invalidTransitionIsTyped(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newAttributeService(d)

	_, err := svc.Create(context.Background(), author(), attribute(), false)
	require.NoError(t, err)

	calls := d.calls
	_, err = svc.SetStatus(context.Background(), author(), "max-age", models.StatusDeployed)
	require.ErrorIs(t, err, ErrInvalidTransition, "PIP data is never deployed through a status change")
	assert.Equal(t, calls, d.calls, "an impossible status is refused as input, before any decision")

	_, err = svc.SetStatus(context.Background(), author(), "max-age", models.StatusAccepted)
	require.NoError(t, err)

	_, err = svc.SetStatus(context.Background(), author(), "max-age", models.StatusAccepted)
	require.ErrorIs(t, err, ErrInvalidTransition, "the store refuses a transition to the same status")

	_, err = svc.SetStatus(context.Background(), author(), "missing", models.StatusAccepted)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestAttributeService_Restore_missingVersionIsNotFound(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newAttributeService(d)

	_, err := svc.Create(context.Background(), author(), attribute(), false)
	require.NoError(t, err)

	_, err = svc.Restore(context.Background(), author(), "max-age", 999)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Equal(t, models.ActionRestore, d.action, "the decision is made before the version is looked up")
}

func TestAttributeService_Delete_decidesDeleteAndHonoursIgnoreMissing(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newAttributeService(d)

	_, err := svc.Delete(context.Background(), author(), "missing", false)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Equal(t, models.ActionDelete, d.action)

	out, err := svc.Delete(context.Background(), author(), "missing", true)
	require.NoError(t, err)
	assert.Equal(t, "missing", out.Key())

	_, err = svc.Create(context.Background(), author(), attribute(), false)
	require.NoError(t, err)

	out, err = svc.Delete(context.Background(), author(), "max-age", false)
	require.NoError(t, err)
	assert.Equal(t, "max-age", out.Key())
}

func TestAttributeService_Delete_deniedNeverTouchesTheStore(t *testing.T) {
	t.Parallel()

	svc, store := newAttributeService(&recorder{})

	_, err := svc.Create(context.Background(), author(), attribute(), false)
	require.NoError(t, err)

	denied := NewAttributeService(slog.New(slog2.NewDummyHandler(slog.LevelDebug)), store, &recorder{deny: true})

	_, err = denied.Delete(context.Background(), author(), "max-age", false)
	require.ErrorIs(t, err, authorization.ErrForbidden)

	stored, _, err := store.GetAttribute("max-age")
	require.NoError(t, err)
	assert.NotNil(t, stored)
}

func TestAttributeService_nilAuthorizerPermits(t *testing.T) {
	t.Parallel()

	svc := NewAttributeService(slog.New(slog2.NewDummyHandler(slog.LevelDebug)), newMemAttributes(), nil)

	_, err := svc.Create(context.Background(), authorization.SystemRequestPrincipal(), attribute(), false)
	require.NoError(t, err)

	list, err := svc.List(context.Background(), authorization.SystemRequestPrincipal())
	require.NoError(t, err)
	assert.Len(t, list, 1)
}
