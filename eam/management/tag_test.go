package management

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// memTags is an in-memory TagStore: the PAP only keeps tags in postgres.
type memTags struct {
	mu      sync.Mutex
	tags    map[string]*oas.Tag
	readErr error
	lastBy  identity.Principal
}

func newMemTags() *memTags { return &memTags{tags: map[string]*oas.Tag{}} }

func (m *memTags) ListTags() (oas.Tags, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make(oas.Tags, 0, len(m.tags))
	for _, t := range m.tags {
		out = append(out, *t)
	}

	return out, nil
}

func (m *memTags) ReadTag(id string) (*oas.Tag, uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.readErr != nil {
		return nil, 0, m.readErr
	}

	return m.tags[id], 1, nil
}

//nolint:gocritic // the store interface passes the principal by value.
func (m *memTags) CreateTag(in *oas.Tag, user identity.Principal) (*oas.Tag, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tags[in.Id], m.lastBy = in, user

	return in, nil
}

//nolint:gocritic // the store interface passes the principal by value.
func (m *memTags) UpdateTag(_ *oas.Tag, _ uint64, in *oas.Tag, user identity.Principal) (*oas.Tag, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tags[in.Id], m.lastBy = in, user

	return in, nil
}

//nolint:gocritic // the store interface passes the principal by value.
func (m *memTags) DeleteTag(prev *oas.Tag, _ uint64, user identity.Principal) (*oas.Tag, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.tags, prev.Id)
	m.lastBy = user

	return prev, nil
}

func newTagService(d *recorder) (*TagService, *memTags) {
	store := newMemTags()
	return NewTagService(slog.New(slog2.NewDummyHandler(slog.LevelDebug)), store, d), store
}

func admin() *authorization.RequestPrincipal {
	return &authorization.RequestPrincipal{Principal: identity.NewPrincipal(identity.KindUser, "bob"), Roles: []string{"admin"}}
}

func tag(id string) *oas.Tag { return &oas.Tag{Id: id, Name: id} }

func TestTagService_List_decidesReadOnTheCollection(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newTagService(d)

	_, err := svc.List(context.Background(), admin())
	require.NoError(t, err)

	assert.Equal(t, 1, d.calls, "one operation is one decision")
	assert.Equal(t, models.ActionRead, d.action)
	assert.Equal(t, "tag::*", d.resource.UID())
}

func TestTagService_Get_deniedBeforeNotFound(t *testing.T) {
	t.Parallel()

	d := &recorder{deny: true}
	svc, _ := newTagService(d)

	_, err := svc.Get(context.Background(), admin(), "missing")
	require.ErrorIs(t, err, authorization.ErrForbidden, "a denied caller must not learn whether the id exists")

	assert.Equal(t, 1, d.calls)
	assert.Equal(t, "tag::missing", d.resource.UID())
}

func TestTagService_Get_notFoundWhenPermitted(t *testing.T) {
	t.Parallel()

	svc, _ := newTagService(&recorder{})

	_, err := svc.Get(context.Background(), admin(), "missing")
	require.ErrorIs(t, err, ErrNotFound)
}

func TestTagService_Get_deniedBeforeTheStoreError(t *testing.T) {
	t.Parallel()

	d := &recorder{deny: true}
	svc, store := newTagService(d)
	store.readErr = errors.New("db down: host pg-1")

	_, err := svc.Get(context.Background(), admin(), "t1")
	require.ErrorIs(t, err, authorization.ErrForbidden, "the store's error text reaches permitted callers only")
}

func TestTagService_Create_decidesCreateAndAttributesTheCaller(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, store := newTagService(d)

	out, err := svc.Create(context.Background(), admin(), tag("laadpalen"), false)
	require.NoError(t, err)
	require.NotNil(t, out)

	assert.Equal(t, models.ActionCreate, d.action)
	assert.Equal(t, "tag::laadpalen", d.resource.UID())
	assert.Equal(t, "bob", store.lastBy.ID)
}

func TestTagService_Create_existingIsConflictAfterTheDecision(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newTagService(d)

	_, err := svc.Create(context.Background(), admin(), tag("t2"), false)
	require.NoError(t, err)

	_, err = svc.Create(context.Background(), admin(), tag("t2"), false)
	require.ErrorIs(t, err, ErrExists)
	assert.Equal(t, 2, d.calls)

	// forceUpsert turns the create into an update of the stored tag.
	_, err = svc.Create(context.Background(), admin(), tag("t2"), true)
	require.NoError(t, err)
	assert.Equal(t, models.ActionUpdate, d.action)
}

func TestTagService_Update_missingIsNotFoundAfterTheDecision(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newTagService(d)

	_, err := svc.Update(context.Background(), admin(), tag("t1"), false)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Equal(t, models.ActionUpdate, d.action)

	// forceUpsert turns the update into a create.
	_, err = svc.Update(context.Background(), admin(), tag("t1"), true)
	require.NoError(t, err)
	assert.Equal(t, models.ActionCreate, d.action)
}

func TestTagService_Delete_missingIsNotFoundUnlessIgnored(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, store := newTagService(d)

	_, err := svc.Delete(context.Background(), admin(), "t1", false)
	require.ErrorIs(t, err, ErrNotFound)
	assert.Equal(t, models.ActionDelete, d.action)

	out, err := svc.Delete(context.Background(), admin(), "t1", true)
	require.NoError(t, err)
	assert.Equal(t, "t1", out.Id)

	_, err = svc.Create(context.Background(), admin(), tag("t1"), false)
	require.NoError(t, err)

	_, err = svc.Delete(context.Background(), admin(), "t1", false)
	require.NoError(t, err)
	assert.Empty(t, store.tags)
}

func TestTagService_Delete_deniedNeverTouchesTheStore(t *testing.T) {
	t.Parallel()

	svc, store := newTagService(&recorder{})
	_, err := svc.Create(context.Background(), admin(), tag("t1"), false)
	require.NoError(t, err)

	denied := NewTagService(slog.New(slog2.NewDummyHandler(slog.LevelDebug)), store, &recorder{deny: true})

	_, err = denied.Delete(context.Background(), admin(), "t1", false)
	require.ErrorIs(t, err, authorization.ErrForbidden)
	assert.Len(t, store.tags, 1)
}

func TestTagService_nilAuthorizerPermits(t *testing.T) {
	t.Parallel()

	svc := NewTagService(slog.New(slog2.NewDummyHandler(slog.LevelDebug)), newMemTags(), nil)

	_, err := svc.Create(context.Background(), authorization.SystemRequestPrincipal(), tag("t1"), false)
	require.NoError(t, err)
}
