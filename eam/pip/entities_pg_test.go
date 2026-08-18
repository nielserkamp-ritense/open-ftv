package pip

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// TestPostgresDB_CreateEntity checks CreateEntity's SQL, field mapping and error handling.
func TestPostgresDB_CreateEntity(t *testing.T) {
	t.Parallel()

	now := time.Date(2025, 8, 13, 12, 10, 8, 123456789, time.UTC)

	u1 := identity.Principal{Kind: identity.KindUser, ID: "sub-123", Name: "donald@duck.us"}
	u2 := identity.Principal{Kind: identity.KindUser, ID: "sub-456", Name: "goofy@weird.us"}

	attrs := models.NewAttributeSet()
	attrs.AddAttributeKV("role", "admin")

	testCases := []struct {
		name    string
		user    identity.Principal
		in      *models.Entity
		wantErr bool
	}{
		{name: "simple insert", user: u1, in: models.NewEntity("user", "u1", nil)},
		{
			name: "full insert",
			user: u2,
			in: models.NewEntity("user", "u2", attrs, "parentA", "parentB").
				WithTitle("title").WithDescription("description").WithTags("x", "y").WithStatus(models.StatusAccepted),
		},
		{name: "force error", user: u1, in: models.NewEntity("user", "u1", nil), wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e, u := tc.in, tc.user

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mock, db := newMockPIP(t, ctx)
			db.now = func() time.Time { return now }

			mock.ExpectBegin()
			mock.ExpectExec("SELECT set_config('openftv.user', $1, true)").WithArgs(u.ID).WillReturnResult(pgxmock.NewResult("SELECT", 1))

			exp := mock.ExpectExec(`INSERT INTO entity
 (status,type,id,title,description,tags,attributes,parents,created,created_by,updated,updated_by)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`).
				WithArgs(e.StatusName(), e.Type(), e.ID(), e.Title(), e.Description(), e.Tags(), gobEncodeAttributes(e.Attributes()), e.Parents(), now, u.ID, now, u.ID)
			if tc.wantErr {
				exp.WillReturnError(errors.New("test error"))
				mock.ExpectRollback()
			} else {
				exp.WillReturnResult(pgxmock.NewResult("CREATE", 1))
				mock.ExpectCommit()
			}

			got, err := db.CreateEntity(ctx, u, e)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, now, got.Created())
				assert.Equal(t, u.ID, got.CreatedBy())
				assert.Equal(t, now, got.Updated())
				assert.Equal(t, u.ID, got.UpdatedBy())
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestPostgresDB_UpdateEntity checks UpdateEntity's SQL, field mapping, concurrency conflict and error handling.
func TestPostgresDB_UpdateEntity(t *testing.T) {
	t.Parallel()

	created := time.Date(2025, 8, 13, 12, 10, 8, 123456789, time.UTC)
	now := created.Add(time.Hour)
	lastIndex := timeToLastIndex(created)

	u1 := identity.Principal{Kind: identity.KindUser, ID: "sub-123", Name: "donald@duck.us"}
	u2 := identity.Principal{Kind: identity.KindUser, ID: "sub-456", Name: "goofy@weird.us"}

	newPrev := func() *models.Entity {
		return models.NewEntity("user", "u1", nil).WithAudit(created, "donald@duck.us", created, "donald@duck.us")
	}

	testCases := []struct {
		name         string
		user         identity.Principal
		in           *models.Entity
		wantErr      bool
		wantMismatch bool
	}{
		{
			name: "simple update",
			user: u1,
			in:   models.NewEntity("user", "u1", nil).WithTitle("new title").WithAudit(created, "donald@duck.us", created, "donald@duck.us"),
		},
		{
			name: "full update",
			user: u2,
			in: models.NewEntity("user", "u1", nil).
				WithTitle("t2").WithDescription("d2").WithTags("a", "b").WithStatus(models.StatusDeployed).
				WithAudit(created, "donald@duck.us", created, "donald@duck.us"),
		},
		{
			name:    "force error",
			user:    u1,
			in:      models.NewEntity("user", "u1", nil).WithAudit(created, "donald@duck.us", created, "donald@duck.us"),
			wantErr: true,
		},
		{
			name:         "force mismatch",
			user:         u1,
			in:           models.NewEntity("user", "u1", nil).WithAudit(created, "donald@duck.us", created, "donald@duck.us"),
			wantMismatch: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prev, e, u := newPrev(), tc.in, tc.user

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mock, db := newMockPIP(t, ctx)
			db.now = func() time.Time { return now }

			mock.ExpectBegin()
			mock.ExpectExec("SELECT set_config('openftv.user', $1, true)").WithArgs(u.ID).WillReturnResult(pgxmock.NewResult("SELECT", 1))

			exp := mock.ExpectExec(`UPDATE entity
 SET title=$4,description=$5,tags=$6,attributes=$7,parents=$8,status=$9,updated=$10,updated_by=$11
 WHERE type=$1 AND id=$2 AND updated=$3`).
				WithArgs(prev.Type(), prev.ID(), created, e.Title(), e.Description(), e.Tags(), gobEncodeAttributes(e.Attributes()), e.Parents(), e.StatusName(), now, u.ID)

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

			got, err := db.UpdateEntity(ctx, u, prev, lastIndex, e)
			if tc.wantErr || tc.wantMismatch {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, created, got.Created())
				assert.Equal(t, "donald@duck.us", got.CreatedBy()) // creator is preserved from prev, not the caller
				assert.Equal(t, now, got.Updated())
				assert.Equal(t, u.ID, got.UpdatedBy())
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
