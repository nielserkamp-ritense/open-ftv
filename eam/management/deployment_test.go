package management

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func newDeploymentService(t *testing.T, d *recorder) (*DeploymentService, *pap.PAP) {
	t.Helper()

	ctx := context.Background()
	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))

	store, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"))
	require.NoError(t, err)

	manager := bundles.NewManager(ctx, logger,
		bundles.WithConfig("../../testdata/unittest/bundles/test2", true),
		bundles.WithPolicyHandler(store))

	return NewDeploymentService(logger, store, manager, d), store
}

func TestDeploymentService_List_decidesReadOnTheCollection(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newDeploymentService(t, d)

	_, err := svc.List(context.Background(), author())
	require.NoError(t, err)

	assert.Equal(t, 1, d.calls, "one operation is one decision")
	assert.Equal(t, models.ActionRead, d.action)
	assert.Equal(t, "deployment::*", d.resource.UID())
}

func TestDeploymentService_Describe_decidesReadOnTheCollection(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newDeploymentService(t, d)

	require.NoError(t, svc.Describe(context.Background(), author()))
	require.NoError(t, svc.Describe(context.Background(), author()))

	assert.Equal(t, 2, d.calls, "every call is decided")
	assert.Equal(t, models.ActionRead, d.action)
	assert.Equal(t, "deployment::*", d.resource.UID())
}

func TestDeploymentService_Get_deniedBeforeNotFound(t *testing.T) {
	t.Parallel()

	d := &recorder{deny: true}
	svc, _ := newDeploymentService(t, d)

	_, err := svc.Get(context.Background(), author(), 42)
	require.ErrorIs(t, err, authorization.ErrForbidden, "a denied caller must not learn whether the version exists")

	assert.Equal(t, "deployment::42", d.resource.UID())
	assert.Nil(t, d.resource.Attributes().GetAttribute(models.AttrStatus), "a missing object contributes its id only")
}

func TestDeploymentService_Get_notFoundWhenPermitted(t *testing.T) {
	t.Parallel()

	svc, _ := newDeploymentService(t, &recorder{})

	_, err := svc.Get(context.Background(), author(), 42)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestDeploymentService_Last_notFoundWhenNothingWasDeployed(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newDeploymentService(t, d)

	_, err := svc.Last(context.Background(), author())
	require.ErrorIs(t, err, ErrNotFound)

	assert.Equal(t, models.ActionRead, d.action)
	assert.Equal(t, "deployment::*", d.resource.UID())
}

func TestDeploymentService_Last_deniedBeforeNotFound(t *testing.T) {
	t.Parallel()

	d := &recorder{deny: true}
	svc, _ := newDeploymentService(t, d)

	_, err := svc.Last(context.Background(), author())
	require.ErrorIs(t, err, authorization.ErrForbidden)
}

func TestDeploymentService_Deploy_decidesDeployOnTheCollection(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, store := newDeploymentService(t, d)

	out, err := svc.Deploy(context.Background(), author(), "Publicatie", "")
	require.NoError(t, err)
	require.NotNil(t, out)

	assert.Equal(t, 1, d.calls, "one operation is one decision")
	assert.Equal(t, models.ActionDeploy, d.action)
	assert.Equal(t, "deployment::*", d.resource.UID())

	last, err := store.LastDeployment()
	require.NoError(t, err)
	require.NotNil(t, last)
	assert.Equal(t, out.Version(), last.Version())

	// the deployment is decided by and read back as the author.
	got, err := svc.Get(context.Background(), author(), out.Version())
	require.NoError(t, err)
	assert.Equal(t, "deployment::"+"1", d.resource.UID())
	assert.NotEmpty(t, d.resource.Attributes().GetAttributeValue(models.AttrStatus), "a stored deployment carries its status")
	assert.Equal(t, "Publicatie", got.ToOAS().Title)
}

func TestDeploymentService_Deploy_deniedTouchesNothing(t *testing.T) {
	t.Parallel()

	d := &recorder{deny: true}
	svc, store := newDeploymentService(t, d)

	_, err := svc.Deploy(context.Background(), &authorization.RequestPrincipal{Principal: identity.NewPrincipal(identity.KindUser, "eve"), Roles: []string{"admin"}}, "Publicatie", "")
	require.ErrorIs(t, err, authorization.ErrForbidden)

	last, _ := store.LastDeployment()
	assert.Nil(t, last, "a denied deploy must NOT create a deployment")
}

func TestDeploymentService_Bundle_unknownIdIsErrBundle(t *testing.T) {
	t.Parallel()

	svc, _ := newDeploymentService(t, &recorder{})

	_, err := svc.Deploy(context.Background(), author(), "Publicatie", "")
	require.NoError(t, err)

	_, err = svc.Bundle(context.Background(), author(), "no-such-bundle")
	require.ErrorIs(t, err, ErrBundle)
}

func TestDeploymentService_withoutAuthorizerPermitsEverything(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
	store, err := pap.New(context.Background(), logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"))
	require.NoError(t, err)

	svc := NewDeploymentService(logger, store, bundles.NewManager(context.Background(), logger), nil)

	_, err = svc.List(context.Background(), authorization.SystemRequestPrincipal())
	require.NoError(t, err)
}
