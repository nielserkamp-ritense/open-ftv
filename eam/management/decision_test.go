package management

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/search"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authlog"
)

// memSearcher is an in-memory search.Searcher that remembers the criteria it was given.
type memSearcher struct{ got *search.Criteria }

func (m *memSearcher) Search(_ context.Context, c *search.Criteria) (oas.AuthlogEntries, error) {
	m.got = c
	return oas.AuthlogEntries{}, nil
}

func TestDecisionService_Search_decidesReadOnTheCollection(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	searcher := &memSearcher{}
	svc := NewDecisionService(searcher, d)

	// criteria that name one entry are still a read of the collection.
	c := &search.Criteria{ID: 42}

	_, err := svc.Search(context.Background(), admin(), c)
	require.NoError(t, err)
	assert.Same(t, c, searcher.got)

	assert.Equal(t, 1, d.calls, "one operation is one decision")
	assert.Equal(t, models.ActionRead, d.action)
	assert.Equal(t, "decision::*", d.resource.UID())
}

func TestDecisionService_Search_deniedNeverSearches(t *testing.T) {
	t.Parallel()

	searcher := &memSearcher{}
	svc := NewDecisionService(searcher, &recorder{deny: true})

	_, err := svc.Search(context.Background(), admin(), &search.Criteria{})
	require.ErrorIs(t, err, authorization.ErrForbidden)
	assert.Nil(t, searcher.got)
}

func TestDecisionService_DecideSearch_decidesWithoutSearching(t *testing.T) {
	t.Parallel()

	d := &recorder{}
	searcher := &memSearcher{}
	svc := NewDecisionService(searcher, d)

	require.NoError(t, svc.DecideSearch(context.Background(), admin()))
	assert.Equal(t, "decision::*", d.resource.UID())
	assert.Nil(t, searcher.got)
}

func TestDecisionService_nilAuthorizerPermits(t *testing.T) {
	t.Parallel()

	svc := NewDecisionService(&memSearcher{}, nil)

	_, err := svc.Search(context.Background(), authorization.SystemRequestPrincipal(), &search.Criteria{})
	require.NoError(t, err)
}
