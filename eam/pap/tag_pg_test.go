package pap

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// TestTagDB_CreateTag checks CreateTag's SQL, field mapping and error handling.
func TestTagDB_CreateTag(t *testing.T) {
	t.Parallel()

	u1 := identity.Principal{Kind: identity.KindUser, ID: "sub-123", Name: "donald@duck.us"}
	u2 := identity.Principal{Kind: identity.KindUser, ID: "sub-456", Name: "goofy@weird.us"}

	testCases := []struct {
		name    string
		user    identity.Principal
		in      *oas.Tag
		wantErr bool
	}{
		{name: "simple insert", user: u1, in: &oas.Tag{Id: "tag1", Name: "Tag One"}},
		{name: "full insert", user: u2, in: &oas.Tag{Id: "tag2", Name: "Tag Two", Description: "a longer description"}},
		{name: "force error", user: u1, in: &oas.Tag{Id: "tag1", Name: "Tag One"}, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tag, u := tc.in, tc.user

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mock, db := newMockTagDB(t, ctx)

			mock.ExpectBegin()
			mock.ExpectExec("SELECT set_config('openftv.user', $1, true)").WithArgs(u.ID).WillReturnResult(pgxmock.NewResult("SELECT", 1))

			exp := mock.ExpectExec(`INSERT INTO tag (tag,title,description,created,created_by,updated,updated_by) VALUES($1,$2,$3,$4,$5,$6,$7)`).
				WithArgs(tag.Id, tag.Name, tag.Description, pgxmock.AnyArg(), u.ID, pgxmock.AnyArg(), u.ID)
			if tc.wantErr {
				exp.WillReturnError(errors.New("test error"))
				mock.ExpectRollback()
			} else {
				exp.WillReturnResult(pgxmock.NewResult("CREATE", 1))
				mock.ExpectCommit()
			}

			got, err := db.CreateTag(ctx, u, tag)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tag.Id, got.Id)
				assert.Equal(t, tag.Name, got.Name)
				assert.Equal(t, tag.Description, got.Description)
				assert.Equal(t, u.ID, got.Audit.CreatedBy.Id)
				assert.Equal(t, u.ID, got.Audit.UpdatedBy.Id)
				assert.Equal(t, got.Audit.Created, got.Audit.Updated)
				_, parseErr := time.Parse(time.RFC3339Nano, got.Audit.Created)
				require.NoError(t, parseErr)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestTagDB_UpdateTag checks UpdateTag's SQL, field mapping, concurrency conflict and error handling.
func TestTagDB_UpdateTag(t *testing.T) {
	t.Parallel()

	created := time.Date(2025, 8, 13, 12, 10, 8, 123456789, time.UTC)
	lastIndex := timeToLastIndex(created)

	u1 := identity.Principal{Kind: identity.KindUser, ID: "sub-123", Name: "donald@duck.us"}
	u2 := identity.Principal{Kind: identity.KindUser, ID: "sub-456", Name: "goofy@weird.us"}

	testCases := []struct {
		name         string
		user         identity.Principal
		prev         *oas.Tag
		in           *oas.Tag
		wantErr      bool
		wantMismatch bool
	}{
		{name: "simple update", user: u1, prev: &oas.Tag{Id: "tag1", Name: "Tag One"}, in: &oas.Tag{Name: "Tag One Renamed"}},
		{
			name: "full update", user: u2,
			prev: &oas.Tag{Id: "tag2", Name: "Tag Two"},
			in:   &oas.Tag{Name: "Tag Two", Description: "new description"},
		},
		{name: "force error", user: u1, prev: &oas.Tag{Id: "tag1", Name: "Tag One"}, in: &oas.Tag{Name: "Tag One"}, wantErr: true},
		{
			name: "force mismatch", user: u1,
			prev: &oas.Tag{Id: "tag1", Name: "Tag One"}, in: &oas.Tag{Name: "Tag One"}, wantMismatch: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prev, tag, u := tc.prev, tc.in, tc.user

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mock, db := newMockTagDB(t, ctx)

			mock.ExpectBegin()
			mock.ExpectExec("SELECT set_config('openftv.user', $1, true)").WithArgs(u.ID).WillReturnResult(pgxmock.NewResult("SELECT", 1))

			exp := mock.ExpectExec(`UPDATE tag SET title=$3,description=$4,updated=$5,updated_by=$6 WHERE tag=$1 AND updated=$2`).
				WithArgs(prev.Id, created, tag.Name, tag.Description, pgxmock.AnyArg(), u.ID)

			switch {
			case tc.wantErr:
				exp.WillReturnError(errors.New("test error"))
				mock.ExpectRollback()
			case tc.wantMismatch:
				exp.WillReturnResult(pgxmock.NewResult("UPDATE", 0))
				mock.ExpectCommit()
			default:
				exp.WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectCommit()
			}

			got, err := db.UpdateTag(ctx, u, prev, lastIndex, tag)

			switch {
			case tc.wantErr:
				require.Error(t, err)
				require.Nil(t, got)
			case tc.wantMismatch:
				require.ErrorIs(t, err, ErrTagConcurrency)
				require.Nil(t, got)
			default:
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tag.Name, got.Name)
				assert.Equal(t, tag.Description, got.Description)
				assert.Equal(t, u.ID, got.Audit.UpdatedBy.Id)
				_, parseErr := time.Parse(time.RFC3339Nano, got.Audit.Updated)
				require.NoError(t, parseErr)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
