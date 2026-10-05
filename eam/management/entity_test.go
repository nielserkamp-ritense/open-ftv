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

func newEntityService(d *recorder) (*EntityService, *memEntities) {
	store := newMemEntities()
	return NewEntityService(slog.New(slog2.NewDummyHandler(slog.LevelDebug)), store, d), store
}

// entity is the submitted form of the entity every test works on, user::alice.
func entity(status string, tags ...string) *oas.Entity {
	return &oas.Entity{Type: "user", Id: "alice", Status: status, Metadata: oas.Metadata{Title: "alice", Tags: tags}}
}

func TestEntityService_List_decidesReadOnTheCollection(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newEntityService(d)

	_, err := svc.List(context.Background(), author())
	require.NoError(t, err)

	assert.Equal(t, 1, d.calls, "one operation is one decision")
	assert.Equal(t, models.ActionRead, d.action)
	assert.Equal(t, "entity::*", d.resource.UID())
}

func TestEntityService_Get_deniedBeforeNotFound(t *testing.T) {
	t.Parallel()

	d := &recorder{deny: true}
	svc, _ := newEntityService(d)

	_, err := svc.Get(context.Background(), author(), "user::missing")
	require.ErrorIs(t, err, authorization.ErrForbidden, "a denied caller must not learn whether the entity exists")

	assert.Equal(t, 1, d.calls)
	assert.Equal(t, "entity::user::missing", d.resource.UID(), "the resource id is the entity's UID")
	assert.Nil(t, d.resource.Attributes().GetAttribute(models.AttrStatus), "a missing object contributes its uid only")
}

func TestEntityService_Get_returnsTheEntityAuditNotTheAttributeAudit(t *testing.T) {
	t.Parallel()

	svc, _ := newEntityService(&recorder{})

	_, err := svc.Create(context.Background(), author(), entity("concept"), false)
	require.NoError(t, err)

	details, err := svc.Get(context.Background(), author(), "user::alice")
	require.NoError(t, err)
	assert.Equal(t, "user::alice", details.Entity.UID())
}

func TestEntityService_Create_decidesCreateWithTheSubmittedProperties(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, store := newEntityService(d)

	out, err := svc.Create(context.Background(), author(), entity("accepted", "laadpalen"), false)
	require.NoError(t, err)
	require.NotNil(t, out)

	assert.Equal(t, models.ActionCreate, d.action)
	assert.Equal(t, "entity::user::alice", d.resource.UID())
	assert.Equal(t, "concept", d.resource.Attributes().GetAttribute(models.AttrStatus).Value(), "a new entity is a concept whatever the body says")
	assert.Equal(t, []string{"laadpalen"}, d.resource.Attributes().GetAttribute(models.AttrTags).Value())

	stored, _, err := store.GetEntity("user::alice")
	require.NoError(t, err)
	assert.Equal(t, models.StatusConcept, stored.Status(), "an entity is never born accepted: that is the accept action")
	assert.Equal(t, "alice", stored.CreatedBy(), "the write is attributed to the caller")
}

func TestEntityService_Create_existingIsConflictAfterTheDecision(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newEntityService(d)

	_, err := svc.Create(context.Background(), author(), entity("concept"), false)
	require.NoError(t, err)

	_, err = svc.Create(context.Background(), author(), entity("concept"), false)
	require.ErrorIs(t, err, ErrExists)
	assert.Equal(t, 2, d.calls, "the conflict is reported after the decision, not instead of it")

	_, err = svc.Create(context.Background(), author(), entity("concept"), true)
	require.NoError(t, err)
	assert.Equal(t, models.ActionUpdate, d.action, "forceUpsert on an existing uid is an update")
}

func TestEntityService_Update_decidesOnTheStoredPropertiesAndStoresAConcept(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, store := newEntityService(d)

	_, err := svc.Create(context.Background(), author(), entity("concept", "laadpalen"), false)
	require.NoError(t, err)

	_, err = store.UpdateEntityStatus("user::alice", models.StatusAccepted, identity.NewSystemPrincipal())
	require.NoError(t, err)

	// the recorder permits what the shipped guard forbids, so this is the service alone.
	out, err := svc.Update(context.Background(), author(), entity("accepted", "other"), false)
	require.NoError(t, err)

	assert.Equal(t, models.ActionUpdate, d.action)
	assert.Equal(t, "accepted", d.resource.Attributes().GetAttribute(models.AttrStatus).Value(), "the PDP sees the stored status, not the submitted one")
	assert.Equal(t, []string{"laadpalen"}, d.resource.Attributes().GetAttribute(models.AttrTags).Value())
	assert.Equal(t, models.StatusConcept, out.Status(), "an edit is a concept whatever the body says: only accept makes it accepted")
}

func TestEntityService_Update_missingIsNotFoundAfterTheDecision(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newEntityService(d)

	_, err := svc.Update(context.Background(), author(), entity("concept"), false)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Equal(t, 1, d.calls)
	assert.Equal(t, models.ActionUpdate, d.action)

	_, err = svc.Update(context.Background(), author(), entity("concept"), true)
	require.NoError(t, err)
	assert.Equal(t, models.ActionCreate, d.action, "forceUpsert on a missing uid is a create")
}

func TestEntityService_SetStatus_isAcceptOrRevert(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newEntityService(d)

	_, err := svc.Create(context.Background(), author(), entity("concept"), false)
	require.NoError(t, err)

	out, err := svc.SetStatus(context.Background(), author(), "user::alice", models.StatusAccepted)
	require.NoError(t, err)
	assert.Equal(t, models.ActionAccept, d.action)
	assert.Equal(t, models.StatusAccepted, out.Status())

	out, err = svc.SetStatus(context.Background(), author(), "user::alice", models.StatusConcept)
	require.NoError(t, err)
	assert.Equal(t, models.ActionRevert, d.action)
	assert.Equal(t, models.StatusConcept, out.Status())

	calls := d.calls
	_, err = svc.SetStatus(context.Background(), author(), "user::alice", models.StatusDeployed)
	require.ErrorIs(t, err, ErrInvalidTransition, "PIP data is never deployed through a status change")
	assert.Equal(t, calls, d.calls, "an impossible status is refused as input, before any decision")
}

func TestEntityService_Restore_missingVersionIsNotFound(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newEntityService(d)

	_, err := svc.Create(context.Background(), author(), entity("concept"), false)
	require.NoError(t, err)

	_, err = svc.Restore(context.Background(), author(), "user::alice", 999)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Equal(t, models.ActionRestore, d.action, "the decision is made before the version is looked up")
}

func TestEntityService_Delete_decidesDeleteAndHonoursIgnoreMissing(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newEntityService(d)

	_, err := svc.Delete(context.Background(), author(), "user::missing", false)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Equal(t, models.ActionDelete, d.action)

	out, err := svc.Delete(context.Background(), author(), "user::missing", true)
	require.NoError(t, err)
	assert.Equal(t, "user::missing", out.UID())

	_, err = svc.Create(context.Background(), author(), entity("concept"), false)
	require.NoError(t, err)

	out, err = svc.Delete(context.Background(), author(), "user::alice", false)
	require.NoError(t, err)
	assert.Equal(t, "user::alice", out.UID())
}

func TestEntityService_nilAuthorizerPermits(t *testing.T) {
	t.Parallel()

	svc := NewEntityService(slog.New(slog2.NewDummyHandler(slog.LevelDebug)), newMemEntities(), nil)

	_, err := svc.Create(context.Background(), authorization.SystemRequestPrincipal(), entity("concept"), false)
	require.NoError(t, err)

	list, err := svc.List(context.Background(), authorization.SystemRequestPrincipal())
	require.NoError(t, err)
	assert.Len(t, list, 1)
}
