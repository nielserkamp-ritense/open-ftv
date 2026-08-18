package principals

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

func newMockPostgres(ctx context.Context, t *testing.T) (pgxmock.PgxPoolIface, *DB) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	pool, err2 := postgresql.NewWithPool(ctx, "postgres://localhost:5432/table", time.Minute, 3, mock)
	require.NoError(t, err2)

	return mock, NewDBWithPool(pool)
}

func TestDB_Upsert_PassesDisplayDataAndNullsUnknownClaims(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		principal identity.Principal
		wantName  any
		wantEmail any
	}{
		{
			name:      "full token",
			principal: identity.Principal{Kind: identity.KindUser, ID: "sub-1", Name: "Ton", Email: "ton@vng.nl", Issuer: "https://idp"},
			wantName:  ptr("Ton"),
			wantEmail: ptr("ton@vng.nl"),
		},
		{
			// A token without the claim must send NULL, so the COALESCE in the statement keeps a
			// name the manager already knew instead of erasing it.
			name:      "token without preferred_username or email",
			principal: identity.Principal{Kind: identity.KindUser, ID: "sub-1", Issuer: "https://idp"},
			wantName:  (*string)(nil),
			wantEmail: (*string)(nil),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mock, db := newMockPostgres(ctx, t)

			mock.ExpectBegin()
			mock.ExpectExec("SELECT set_config").WithArgs(pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("SELECT", 1))
			mock.ExpectExec("INSERT INTO principal").
				WithArgs(tc.principal.ID, KindUser, tc.principal.Issuer, tc.wantName, tc.wantEmail).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))
			mock.ExpectCommit()

			require.NoError(t, db.Upsert(ctx, tc.principal))
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDB_Upsert_WrapsFailure(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mock, db := newMockPostgres(ctx, t)

	boom := errors.New("read-only transaction")

	mock.ExpectBegin()
	mock.ExpectExec("SELECT set_config").WithArgs(pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("SELECT", 1))
	mock.ExpectExec("INSERT INTO principal").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(boom)
	mock.ExpectRollback()

	err := db.Upsert(ctx, identity.Principal{Kind: identity.KindUser, ID: "sub-1"})
	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
}

func TestDB_Resolve_ReturnsOnlyKnownIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mock, db := newMockPostgres(ctx, t)

	rows := pgxmock.NewRows([]string{"id", "kind", "display_name", "email"}).
		AddRow("sub-1", KindUser, "Ton", "ton@vng.nl").
		AddRow("*SEED*", KindSystem, "*SEED*", "")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, kind, display_name, email FROM principal").
		WithArgs([]string{"sub-1", "*SEED*", "sub-unseen"}).
		WillReturnRows(rows)
	mock.ExpectCommit()

	got, err := db.Resolve(ctx, []string{"sub-1", "*SEED*", "sub-unseen", "sub-1"})
	require.NoError(t, err)

	// The unknown id is simply absent, which is what makes the caller fall back to the raw id.
	assert.Len(t, got, 2)
	assert.Equal(t, Record{ID: "sub-1", Kind: KindUser, Name: "Ton", Email: "ton@vng.nl"}, got["sub-1"])
	assert.NotContains(t, got, "sub-unseen")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_Resolve_NoIDsSkipsTheQuery(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mock, db := newMockPostgres(ctx, t)

	got, err := db.Resolve(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, got)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDedupe(t *testing.T) {
	t.Parallel()

	// A list view of many objects written by two people must ask for two ids.
	assert.Equal(t, []string{"a", "b"}, dedupe([]string{"a", "b", "a", "", "b"}))
	assert.Empty(t, dedupe([]string{"", ""}))
}

func TestRecord_Linkable(t *testing.T) {
	t.Parallel()

	// Only a real subject can be linked to: a legacy row is keyed by a display name and
	// identifies nobody, and a system row is the manager itself.
	assert.True(t, Record{Kind: KindUser}.Linkable())
	assert.False(t, Record{Kind: KindLegacy}.Linkable())
	assert.False(t, Record{Kind: KindSystem}.Linkable())
}

func ptr(s string) *string { return &s }

func TestDB_Upsert_PreservesIssuerWhenTheRequestCarriedNoJWT(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mock, db := newMockPostgres(ctx, t)

	mock.ExpectBegin()
	mock.ExpectExec("SELECT set_config").WithArgs(pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("SELECT", 1))
	// A principal can be resolved from a `user::<id>` attribute with no token behind it, so the
	// statement must not overwrite a known issuer with the empty string.
	mock.ExpectExec(regexp.QuoteMeta("COALESCE(NULLIF(EXCLUDED.issuer, ''), principal.issuer)")).
		WithArgs("sub-1", KindUser, "", (*string)(nil), (*string)(nil)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	require.NoError(t, db.Upsert(ctx, identity.Principal{Kind: identity.KindUser, ID: "sub-1"}))
	require.NoError(t, mock.ExpectationsWereMet())
}
