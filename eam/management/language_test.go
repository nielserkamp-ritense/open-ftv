package management

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// memLanguages is an in-memory LanguageStore.
type memLanguages struct{ calls int }

func (m *memLanguages) ListLanguages() (oas.Languages, error) {
	m.calls++
	return oas.Languages{{Id: "cedar"}}, nil
}

func TestLanguageService_List_decidesReadOnTheCollection(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	svc := NewLanguageService(&memLanguages{}, d)

	out, err := svc.List(context.Background(), admin())
	require.NoError(t, err)
	assert.Len(t, out, 1)

	assert.Equal(t, 1, d.calls, "one operation is one decision")
	assert.Equal(t, models.ActionRead, d.action)
	assert.Equal(t, "language::*", d.resource.UID())
}

func TestLanguageService_List_deniedNeverTouchesTheStore(t *testing.T) {
	t.Parallel()

	store := &memLanguages{}
	svc := NewLanguageService(store, &recorder{deny: true})

	_, err := svc.List(context.Background(), admin())
	require.ErrorIs(t, err, authorization.ErrForbidden)
	assert.Zero(t, store.calls)
}

func TestLanguageService_nilAuthorizerPermits(t *testing.T) {
	t.Parallel()

	svc := NewLanguageService(&memLanguages{}, nil)

	_, err := svc.List(context.Background(), authorization.SystemRequestPrincipal())
	require.NoError(t, err)
}
