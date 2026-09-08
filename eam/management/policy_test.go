package management

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/kvtools/valkeyrie/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

// recorder is a decider that records the question and answers as told.
type recorder struct {
	deny     bool
	caller   *authorization.RequestPrincipal
	action   string
	resource *models.Entity
	calls    int
}

func (r *recorder) Decide(_ context.Context, caller *authorization.RequestPrincipal, action string, resource *models.Entity) error {
	r.calls++
	r.caller, r.action, r.resource = caller, action, resource

	if r.deny {
		return authorization.ErrForbidden
	}

	return nil
}

func newPolicyService(t *testing.T, d *recorder) (*PolicyService, *pap.PAP) {
	t.Helper()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
	store, err := pap.New(context.Background(), logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"))
	require.NoError(t, err)

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

	out, err := svc.Create(context.Background(), author(), policy(t, "p1", "laadpalen"), nil, false)
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

// An app caller is decided as itself (so `principal is app` rules apply) but attributed to the
// system principal: its id is the API key, and no principal row exists for it.
func TestPolicyService_Create_appCallerIsAttributedToTheSystem(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, store := newPolicyService(t, d)

	app := &authorization.RequestPrincipal{Principal: identity.NewPrincipal(identity.KindApp, "s3cret-key")}

	_, err := svc.Create(context.Background(), app, policy(t, "p1"), nil, false)
	require.NoError(t, err)

	assert.Equal(t, identity.KindApp, d.caller.Kind, "the PDP sees the app, so `principal is app` rules apply")
	assert.Equal(t, identity.NewSystemPrincipal(), d.caller.Attribution(), "the store never sees the API key")

	stored, _, err := store.Read("p1")
	require.NoError(t, err)
	require.NotNil(t, stored)
}

func TestPolicyService_Create_existingIsConflictAfterTheDecision(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newPolicyService(t, d)

	_, err := svc.Create(context.Background(), author(), policy(t, "p2"), nil, false)
	require.NoError(t, err)

	_, err = svc.Create(context.Background(), author(), policy(t, "p2"), nil, false)
	require.ErrorIs(t, err, ErrExists)
	assert.Equal(t, 2, d.calls)

	// forceUpsert turns the create into an update of the stored object.
	_, err = svc.Create(context.Background(), author(), policy(t, "p2"), nil, true)
	require.NoError(t, err)
	assert.Equal(t, models.ActionUpdate, d.action)
}

func TestPolicyService_Update_missingIsNotFoundAfterTheDecision(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newPolicyService(t, d)

	_, err := svc.Update(context.Background(), author(), policy(t, "p1"), nil, false)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Equal(t, models.ActionUpdate, d.action)

	// forceUpsert turns the update into a create.
	_, err = svc.Update(context.Background(), author(), policy(t, "p1"), nil, true)
	require.NoError(t, err)
	assert.Equal(t, models.ActionCreate, d.action)
}

func TestPolicyService_SetStatus_isAcceptOrDeploy(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newPolicyService(t, d)

	_, err := svc.Create(context.Background(), author(), policy(t, "p1"), nil, false)
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

	_, err := svc.Create(context.Background(), author(), policy(t, "p1"), nil, false)
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
	store, err := pap.New(context.Background(), logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"))
	require.NoError(t, err)

	svc := NewPolicyService(logger, store, nil)

	_, err = svc.List(context.Background(), authorization.SystemRequestPrincipal())
	require.NoError(t, err)
}

// A store that fails to read is still decided, and a denied caller gets the denial: the
// store's state and error text reach permitted callers only.
func TestPolicyService_readFailureIsDecidedBeforeItIsReported(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
	store, err := pap.New(context.Background(), logger, pap.WithKeyValueDB(&failingStore{Store: memory.New()}, ""), pap.WithLanguage("cedar"))
	require.NoError(t, err)

	t.Run("denied caller sees the denial", func(t *testing.T) {
		d := &recorder{deny: true}
		svc := NewPolicyService(logger, store, d)

		_, err := svc.Get(context.Background(), author(), "p1")
		require.ErrorIs(t, err, authorization.ErrForbidden)
		assert.Equal(t, models.ActionRead, d.action)
		assert.Equal(t, "policy::p1", d.resource.UID())
		assert.Nil(t, d.resource.Attributes().GetAttributeValue(models.AttrStatus), "an unread object contributes its id only")

		_, err = svc.Delete(context.Background(), author(), "p1", false)
		require.ErrorIs(t, err, authorization.ErrForbidden)
		assert.Equal(t, models.ActionDelete, d.action)
	})

	t.Run("permitted caller sees the store error", func(t *testing.T) {
		d := &recorder{}
		svc := NewPolicyService(logger, store, d)

		_, err := svc.Get(context.Background(), author(), "p1")
		require.ErrorIs(t, err, errStoreDown)
		assert.Equal(t, 1, d.calls)
	})
}

// The uri a policy points at is fetched only once the caller is permitted.
func TestPolicyService_Create_resolvesContentAfterTheDecision(t *testing.T) {
	t.Parallel()

	t.Run("denied", func(t *testing.T) {
		t.Parallel()

		d := &recorder{deny: true}
		svc, _ := newPolicyService(t, d)

		fetched := false
		resolve := func(context.Context) (*models.Policy, error) {
			fetched = true
			return policy(t, "p1"), nil
		}

		_, err := svc.Create(context.Background(), author(), policy(t, "p1"), resolve, false)
		require.ErrorIs(t, err, authorization.ErrForbidden)
		assert.False(t, fetched, "a denied caller must not make the manager fetch")
	})

	t.Run("permitted", func(t *testing.T) {
		t.Parallel()

		d := &recorder{}
		svc, store := newPolicyService(t, d)

		resolve := func(context.Context) (*models.Policy, error) {
			return policy(t, "p1").WithTitle("fetched"), nil
		}

		out, err := svc.Create(context.Background(), author(), policy(t, "p1"), resolve, false)
		require.NoError(t, err)
		assert.Equal(t, "fetched", out.ToOAS(false).Metadata.Title)

		stored, _, err := store.Read("p1")
		require.NoError(t, err)
		assert.Equal(t, "fetched", stored.ToOAS(false).Metadata.Title)
	})

	t.Run("unresolvable", func(t *testing.T) {
		t.Parallel()

		d := &recorder{}
		svc, store := newPolicyService(t, d)

		resolve := func(context.Context) (*models.Policy, error) { return nil, errors.New("connection refused") }

		_, err := svc.Update(context.Background(), author(), policy(t, "p1"), resolve, true)
		require.ErrorIs(t, err, ErrContent)

		stored, _, err := store.Read("p1")
		require.NoError(t, err)
		assert.Nil(t, stored, "nothing is stored when the content cannot be resolved")
	})
}

// A restore of a version the store does not hold is ErrNotFound, not a panic.
func TestPolicyService_Restore_missingVersionIsNotFound(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newPolicyService(t, d)

	_, err := svc.Create(context.Background(), author(), policy(t, "p1"), nil, false)
	require.NoError(t, err)

	_, err = svc.Restore(context.Background(), author(), "p1", 999)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Equal(t, models.ActionRestore, d.action, "the decision is made before the version is looked up")
}

var errStoreDown = errors.New("store down")

// failingStore is a key-value store whose reads fail.
type failingStore struct {
	store.Store
}

func (f *failingStore) Get(context.Context, string, *store.ReadOptions) (*store.KVPair, error) {
	return nil, errStoreDown
}
