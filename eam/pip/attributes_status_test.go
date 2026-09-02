package pip

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

// testAttributeDB is an AttributePersister test double returning canned data.
type testAttributeDB struct {
	AttributePersister
	prev      *models.Attribute
	readErr   error
	updateErr error

	updateCalls int
	updateUser  identity.Principal
	updated     *models.Attribute
}

func (d *testAttributeDB) ReadAttribute(context.Context, string) (*models.Attribute, uint64, error) {
	return d.prev, 1, d.readErr
}

func (d *testAttributeDB) UpdateAttribute(_ context.Context, user identity.Principal, _ *models.Attribute, _ uint64, a *models.Attribute) (*models.Attribute, error) {
	d.updateCalls++
	d.updateUser = user
	d.updated = a

	if d.updateErr != nil {
		return nil, d.updateErr
	}

	return a, nil
}

// TestPIP_UpdateAttributeStatus checks the allowed status transitions, error paths, and the persisted Principal.
func TestPIP_UpdateAttributeStatus(t *testing.T) {
	t.Parallel()

	user := identity.Principal{Kind: identity.KindUser, ID: "sub-123", Name: "donald@duck.us"}

	testCases := []struct {
		name            string
		prevStatus      models.Status // ignored when notFound is set
		notFound        bool
		readErr         error
		newStatus       models.Status
		updateErr       error
		wantErr         bool
		wantUpdateCalls int
	}{
		{name: "concept to accepted is allowed", prevStatus: models.StatusConcept, newStatus: models.StatusAccepted, wantUpdateCalls: 1},
		{name: "accepted to concept is allowed", prevStatus: models.StatusAccepted, newStatus: models.StatusConcept, wantUpdateCalls: 1},
		{name: "concept to concept is rejected", prevStatus: models.StatusConcept, newStatus: models.StatusConcept, wantErr: true},
		{name: "accepted to accepted is rejected", prevStatus: models.StatusAccepted, newStatus: models.StatusAccepted, wantErr: true},
		{name: "concept to deployed is rejected", prevStatus: models.StatusConcept, newStatus: models.StatusDeployed, wantErr: true},
		{name: "deployed to accepted is rejected", prevStatus: models.StatusDeployed, newStatus: models.StatusAccepted, wantErr: true},
		{name: "deployed to concept is rejected", prevStatus: models.StatusDeployed, newStatus: models.StatusConcept, wantErr: true},
		{name: "unknown attribute", notFound: true, newStatus: models.StatusAccepted, wantErr: true},
		{name: "read error propagates", prevStatus: models.StatusConcept, readErr: errors.New("db down"), newStatus: models.StatusAccepted, wantErr: true},
		{
			name: "persist error propagates", prevStatus: models.StatusConcept, newStatus: models.StatusAccepted,
			updateErr: errors.New("db down"), wantErr: true, wantUpdateCalls: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := &testAttributeDB{readErr: tc.readErr, updateErr: tc.updateErr}
			if !tc.notFound {
				db.prev = models.NewAttribute("k1", "v1").WithStatus(tc.prevStatus)
			}

			p, err := New(context.Background(), slog.New(util.NewDummyHandler(slog.LevelInfo)), WithKeyValueDB(memory.New(), ""))
			require.NoError(t, err)
			p.attributeDB = db

			got, err := p.UpdateAttributeStatus("k1", tc.newStatus, user)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tc.newStatus, got.Status())
			}

			// db.updateCalls distinguishes short-circuited rejections (0) from an attempted persist (1),
			// whether that attempt succeeded or failed.
			assert.Equal(t, tc.wantUpdateCalls, db.updateCalls)

			if tc.wantUpdateCalls > 0 {
				assert.Equal(t, user, db.updateUser)
			}
		})
	}
}
