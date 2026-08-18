package authorization

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// recorderStub captures what recordPrincipal handed to the store, and can fail on demand.
type recorderStub struct {
	seen []identity.Principal
	err  error
}

func (r *recorderStub) Upsert(_ context.Context, p identity.Principal) error {
	r.seen = append(r.seen, p)
	return r.err
}

func newAuthWithRecorder(rec PrincipalRecorder) *auth {
	return &auth{
		ctx:        context.Background(),
		log:        slog.New(slog2.NewDummyHandler(slog.LevelInfo)),
		principals: rec,
	}
}

func TestAuth_recordPrincipal_OnlyAuthenticatedUsers(t *testing.T) {
	t.Parallel()

	user := identity.Principal{Kind: identity.KindUser, ID: "sub-123", Name: "Ton"}

	testCases := []struct {
		name string
		p    identity.Principal
		want bool
	}{
		{name: "authenticated user is recorded", p: user, want: true},
		// The manager's own actions are already seeded as principals; recording them on every
		// request would rewrite seeded rows with request data.
		{name: "system principal is not recorded", p: identity.NewSystemPrincipal()},
		{name: "unknown principal is not recorded", p: identity.NewUnknownPrincipal()},
		{name: "user without an id is not recorded", p: identity.Principal{Kind: identity.KindUser}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec := &recorderStub{}
			a := newAuthWithRecorder(rec)

			require.NoError(t, a.recordPrincipal(http.MethodGet, tc.p))

			if tc.want {
				require.Len(t, rec.seen, 1)
				assert.Equal(t, tc.p, rec.seen[0])
			} else {
				assert.Empty(t, rec.seen)
			}
		})
	}
}

func TestAuth_recordPrincipal_FailureIsFatalOnlyForWrites(t *testing.T) {
	t.Parallel()

	user := identity.Principal{Kind: identity.KindUser, ID: "sub-123"}
	boom := errors.New("database is read-only")

	testCases := []struct {
		name    string
		method  string
		wantErr bool
	}{
		// A degraded database must never lock everyone out of the management UI, so reads carry on.
		{name: "GET tolerates a failed upsert", method: http.MethodGet},
		{name: "HEAD tolerates a failed upsert", method: http.MethodHead},
		{name: "OPTIONS tolerates a failed upsert", method: http.MethodOptions},
		// created_by is a foreign key to principal, so a write without the row would fail anyway,
		// deep inside the handler and as an unhelpful foreign-key violation.
		{name: "POST fails", method: http.MethodPost, wantErr: true},
		{name: "PUT fails", method: http.MethodPut, wantErr: true},
		{name: "DELETE fails", method: http.MethodDelete, wantErr: true},
		{name: "lowercase method is still a write", method: "post", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := newAuthWithRecorder(&recorderStub{err: boom})

			err := a.recordPrincipal(tc.method, user)
			if !tc.wantErr {
				assert.NoError(t, err)
				return
			}

			var target *ErrPrincipalNotRecorded

			require.ErrorAs(t, err, &target, "must be typed so handlers answer 500 rather than 403")
			assert.ErrorIs(t, err, boom)
		})
	}
}

func TestAuth_recordPrincipal_NoStoreConfigured(t *testing.T) {
	t.Parallel()

	a := newAuthWithRecorder(nil)

	// Without postgres there is no principal table, and no foreign key that a write could violate.
	assert.NoError(t, a.recordPrincipal(http.MethodPost, identity.Principal{Kind: identity.KindUser, ID: "sub-123"}))
}
