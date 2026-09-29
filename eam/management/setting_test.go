package management

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/settings"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/settings"
)

// memSettings is an in-memory settings.SettingsPersister.
type memSettings struct {
	saved  *oas.Settings
	lastBy identity.Principal
}

func (m *memSettings) GetSettings(context.Context) (*oas.Settings, error) {
	if m.saved == nil {
		return settings.Defaults(), nil
	}

	return m.saved, nil
}

//nolint:gocritic // the store interface passes the principal by value.
func (m *memSettings) UpdateSettings(_ context.Context, user identity.Principal, in *oas.Settings) (*oas.Settings, error) {
	m.saved, m.lastBy = in, user
	return in, nil
}

func newSettingService(d *recorder) (*SettingService, *memSettings) {
	store := &memSettings{}
	return NewSettingService(store, d), store
}

func TestSettingService_Get_decidesReadOnTheUISettings(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, _ := newSettingService(d)

	out, err := svc.Get(context.Background(), admin())
	require.NoError(t, err)
	assert.Equal(t, settings.DefaultHeaderTitle, out.HeaderTitle, "nothing saved yet answers with the defaults")

	assert.Equal(t, 1, d.calls, "one operation is one decision")
	assert.Equal(t, models.ActionRead, d.action)
	assert.Equal(t, "setting::ui", d.resource.UID())
}

func TestSettingService_Get_denied(t *testing.T) {
	t.Parallel()

	svc, _ := newSettingService(&recorder{deny: true})

	_, err := svc.Get(context.Background(), admin())
	require.ErrorIs(t, err, authorization.ErrForbidden)
}

func TestSettingService_Update_decidesUpdateAndAttributesTheCaller(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc, store := newSettingService(d)

	_, err := svc.Update(context.Background(), admin(), &oas.Settings{HeaderTitle: "ACME"})
	require.NoError(t, err)

	assert.Equal(t, models.ActionUpdate, d.action, "the record always exists to the caller: never a create")
	assert.Equal(t, "setting::ui", d.resource.UID())
	assert.Equal(t, "bob", store.lastBy.ID)
	assert.Equal(t, "ACME", store.saved.HeaderTitle)
}

func TestSettingService_Update_deniedNeverTouchesTheStore(t *testing.T) {
	t.Parallel()

	svc, store := newSettingService(&recorder{deny: true})

	_, err := svc.Update(context.Background(), admin(), &oas.Settings{HeaderTitle: "ACME"})
	require.ErrorIs(t, err, authorization.ErrForbidden)
	assert.Nil(t, store.saved)
}

func TestSettingService_nilAuthorizerPermits(t *testing.T) {
	t.Parallel()

	svc := NewSettingService(&memSettings{}, nil)

	_, err := svc.Update(context.Background(), authorization.SystemRequestPrincipal(), &oas.Settings{HeaderTitle: "ACME"})
	require.NoError(t, err)
}
