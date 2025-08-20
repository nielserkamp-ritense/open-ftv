package pap

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

func TestTimeToLastIndex(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   time.Time
		want uint64
	}{
		{
			name: "zero",
		},
		{
			name: "1 jan 1970 00:00:00.000000001",
			in:   time.Date(1970, 1, 1, 0, 0, 0, 1, time.UTC),
			want: 1,
		},
		{
			name: "1 jan 2020 00:00:00.000000000",
			in:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			want: 0x15e59a35b98a0000,
		},
		{
			name: "13 august 2025 12:10:08.123456789",
			in:   time.Date(2025, 8, 13, 12, 10, 8, 123456789, time.UTC),
			want: 0x185b5255c51a0d15,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := timeToLastIndex(tc.in)
			assert.Equal(t, tc.want, got)

			got2 := timeFromLastIndex(got)
			assert.True(t, tc.in.Equal(got2))
		})
	}
}

func TestPolicyFromValues_Fail(t *testing.T) {
	t.Parallel()

	t.Run("policy from values - fail", func(t *testing.T) {
		t.Parallel()

		got, err := policyFromValues([]any{1, 2, 3})
		require.Error(t, err)
		assert.Nil(t, got)
	})
}

func TestNewPostgresDB(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		dsn     string
		maxLife time.Duration
		maxConn int32
		wantErr bool
	}{
		{name: "bad dsn", dsn: "hello world", maxLife: 10 * time.Minute, maxConn: 5, wantErr: true},
		{name: "good dsn", dsn: "postgres://localhost:5432/myDB?sslmode=disable", maxLife: 10 * time.Minute, maxConn: 5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			got, err := NewPostgresDB(ctx, tc.dsn, tc.maxLife, tc.maxConn)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
			}
		})
	}
}

func TestNewPostgresWithPool(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		dsn     string
		maxLife time.Duration
		maxConn int32
		wantErr bool
	}{
		{name: "bad dsn", dsn: "hello world", maxLife: 10 * time.Minute, maxConn: 5, wantErr: true},
		{name: "good dsn", dsn: "postgres://localhost:5432/myDB?sslmode=disable", maxLife: 10 * time.Minute, maxConn: 5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			pool, err := postgresql.New(ctx, tc.dsn, tc.maxLife, tc.maxConn)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, pool)
			} else {
				require.NoError(t, err)
				require.NotNil(t, pool)

				got := NewPostgresWithPool(pool)
				require.NotNil(t, got)
				assert.Equal(t, pool, got.p)
			}
		})
	}
}

func TestPostgresDB_CreatePolicy(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	u1 := "donald@duck.us"
	u2 := "goofy@weird.us"

	p1, e1 := models.NewPolicyFromData("p1", "cedar", "rvva1", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, e1)
	p2, e2 := models.NewPolicyFromData("p2", "opa", "rvva2", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, e2)

	p1.WithTitle("title1")
	p2.WithTags("x", "y", "z").WithTitle("title2").WithDescription("description2")

	wp1, we1 := models.NewPolicyFromData("p1", "cedar", "rvva1", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, we1)
	wp2, we2 := models.NewPolicyFromData("p2", "opa", "rvva2", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, we2)

	wp1.WithTitle("title1").WithAudit(now, u1, now, u1)
	wp2.WithTags("x", "y", "z").WithTitle("title2").WithDescription("description2").WithAudit(now, u2, now, u2)

	testCases := []struct {
		name    string
		user    string
		in      *models.Policy
		wantErr bool
		want    *models.Policy
	}{
		{name: "simple insert", user: u1, in: p1, want: wp1},
		{name: "full insert", user: u2, in: p2, want: wp2},
		{name: "force error", user: u1, in: p1, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p, u := tc.in, tc.user

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mock, db := newMockPG(t, ctx, "postgres://localhost:5432/myDB", time.Minute, 3, now)

			mock.ExpectBegin()

			exp := mock.ExpectExec(`INSERT INTO policy
 (language,id,title,description,rvva_id,uri,tags,content,created,created_by,updated,updated_by)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`).
				WithArgs(p.Language(), p.ID(), p.Title(), p.Description(), p.RvvaID(), p.URI(), p.Tags(), p.ContentString(), now, u, now, u)
			if tc.wantErr {
				exp.WillReturnError(errors.New("test error"))
				mock.ExpectRollback()
			} else {
				exp.WillReturnResult(pgxmock.NewResult("CREATE", 1))
				mock.ExpectCommit()
			}

			got, err := db.CreatePolicy(ctx, u, p)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				policiesMatch(t, tc.want, got)
			}
		})
	}
}

func TestPostgresDB_ReadPolicy(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	u1 := "donald@duck.us"
	u2 := "goofy@weird.us"

	wp1, we1 := models.NewPolicyFromData("73bf13d7-ac8a-488e-b7c7-c5f53cfccb54", "cedar", "rvva1", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, we1)
	wp2, we2 := models.NewPolicyFromData("ec3cc76b-38d4-4819-88ff-5ac900640bca", "opa", "rvva2", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, we2)

	wp1.WithAudit(now, u1, now, u1)
	wp2.WithTags("x", "y", "z").WithTitle("title2").WithDescription("description2").WithAudit(now, u2, now, u2)

	testCases := []struct {
		name     string
		language string
		id       string
		wantErr  bool
		wantIX   uint64
		want     *models.Policy
	}{
		{name: "policy 1", language: "cedar", id: "73bf13d7-ac8a-488e-b7c7-c5f53cfccb54", wantIX: timeToLastIndex(now), want: wp1},
		{name: "policy 2", language: "opa", id: "ec3cc76b-38d4-4819-88ff-5ac900640bca", wantIX: timeToLastIndex(now), want: wp2},
		{name: "force error", language: "cedar", id: "73bf13d7-ac8a-488e-b7c7-c5f53cfccb54", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			id := tc.id
			w := tc.want

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mock, db := newMockPG(t, ctx, "postgres://localhost:5432/myDB", time.Minute, 3, now)

			mock.ExpectBegin()

			exp := mock.ExpectQuery(`SELECT language,id,title,description,rvva_id,uri,tags,content,created,created_by,updated,updated_by
 FROM policy
 WHERE id=$1`).WithArgs(id)
			if tc.wantErr {
				exp.WillReturnError(errors.New("test error"))
				mock.ExpectRollback()
			} else {
				exp.WillReturnRows(
					pgxmock.NewRows([]string{"language", "id", "title", "description", "rvva_id", "uri", "tags", "content", "created", "created_by", "updated", "updated_by"}).
						AddRow(w.Language(), w.ID(), w.Title(), w.Description(), w.RvvaID(), w.URI(), w.Tags(), w.ContentString(), w.Created(), w.CreatedBy(), w.Updated(), w.UpdatedBy()))
				mock.ExpectCommit()
			}

			got, gotIX, err := db.ReadPolicy(ctx, id)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
				assert.Zero(t, gotIX)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantIX, gotIX)
				policiesMatch(t, tc.want, got)
			}
		})
	}
}

func TestPostgresDB_UpdatePolicy(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	u1 := "donald@duck.us"
	u2 := "goofy@weird.us"

	p1, e1 := models.NewPolicyFromData("p1", "cedar", "rvva1", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, e1)
	p2, e2 := models.NewPolicyFromData("p2", "opa", "rvva2", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, e2)

	p1.WithAudit(now, u2, now, u2)
	p2.WithTags("x", "y", "z").WithTitle("title2").WithDescription("description2").WithAudit(now, u1, now, u1)

	wp1, we1 := models.NewPolicyFromData("p1", "cedar", "rvva1", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, we1)
	wp2, we2 := models.NewPolicyFromData("p2", "opa", "rvva2", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, we2)

	wp1.WithAudit(now, u2, now, u1)
	wp2.WithTags("x", "y", "z").WithTitle("title2").WithDescription("description2").WithAudit(now, u1, now, u2)

	testCases := []struct {
		name         string
		user         string
		in           *models.Policy
		wantErr      bool
		wantMismatch bool
		want         *models.Policy
	}{
		{name: "simple update", user: u1, in: p1, want: wp1},
		{name: "full update", user: u2, in: p2, want: wp2},
		{name: "force error", user: u1, in: p1, wantErr: true},
		{name: "force mismatch", user: u1, in: p1, wantMismatch: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p, u := tc.in, tc.user

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mock, db := newMockPG(t, ctx, "postgres://localhost:5432/myDB", time.Minute, 3, now)

			mock.ExpectBegin()

			exp := mock.ExpectExec(`UPDATE policy
 SET language=$3,title=$4,description=$5,rvva_id=$6,uri=$7,tags=$8,content=$9,updated=$10,updated_by=$11
 WHERE id=$1 AND updated=$2`).
				WithArgs(p.ID(), now, p.Language(), p.Title(), p.Description(), p.RvvaID(), p.URI(), p.Tags(), p.ContentString(), now, u)
			if tc.wantErr {
				exp.WillReturnError(errors.New("test error"))
				mock.ExpectRollback()
			} else {
				if tc.wantMismatch {
					exp.WillReturnResult(pgxmock.NewResult("UPDATE", 0))
				} else {
					exp.WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				}
				mock.ExpectCommit()
			}

			got, err := db.UpdatePolicy(ctx, u, p, timeToLastIndex(now), p)
			if tc.wantErr || tc.wantMismatch {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				policiesMatch(t, tc.want, got)
			}
		})
	}
}

func TestPostgresDB_DeletePolicy(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	u1 := "donald@duck.us"
	u2 := "goofy@weird.us"

	p1, e1 := models.NewPolicyFromData("p1", "cedar", "rvva1", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, e1)
	p2, e2 := models.NewPolicyFromData("p2", "opa", "rvva2", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, e2)

	p1.WithAudit(now, u1, now, u1)
	p2.WithTags("x", "y", "z").WithTitle("title2").WithDescription("description2").WithAudit(now, u2, now, u2)

	wp1, we1 := models.NewPolicyFromData("p1", "cedar", "rvva1", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, we1)
	wp2, we2 := models.NewPolicyFromData("p2", "opa", "rvva2", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, we2)

	wp1.WithAudit(now, u1, now, u1)
	wp2.WithTags("x", "y", "z").WithTitle("title2").WithDescription("description2").WithAudit(now, u2, now, u2)

	testCases := []struct {
		name         string
		user         string
		in           *models.Policy
		wantErr      bool
		wantMismatch bool
		want         *models.Policy
	}{
		{name: "policy 1", user: u1, in: p1, want: wp1},
		{name: "policy 2", user: u2, in: p2, want: wp2},
		{name: "force error", user: u1, in: p1, wantErr: true},
		{name: "force mismatch", user: u1, in: p1, wantMismatch: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p, u := tc.in, tc.user

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mock, db := newMockPG(t, ctx, "postgres://localhost:5432/myDB", time.Minute, 3, now)

			mock.ExpectBegin()

			exp := mock.ExpectExec(`DELETE policy
 WHERE id=$1 AND updated=$2`).WithArgs(p.ID(), now)
			if tc.wantErr {
				exp.WillReturnError(errors.New("test error"))
				mock.ExpectRollback()
			} else {
				if tc.wantMismatch {
					exp.WillReturnResult(pgxmock.NewResult("DELETE", 0))
				} else {
					exp.WillReturnResult(pgxmock.NewResult("DELETE", 1))
				}
				mock.ExpectCommit()
			}

			got, err := db.DeletePolicy(ctx, u, p, timeToLastIndex(now))
			if tc.wantErr || tc.wantMismatch {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				policiesMatch(t, tc.want, got)
			}
		})
	}
}

func TestPostgresDB_ListPolicies(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	u1 := "donald@duck.us"
	u2 := "goofy@weird.us"

	wp1, we1 := models.NewPolicyFromData("p1", "cedar", "rvva1", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, we1)
	wp2, we2 := models.NewPolicyFromData("p2", "opa", "rvva2", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, we2)

	wp1.WithAudit(now, u1, now, u1)
	wp2.WithTags("x", "y", "z").WithTitle("title2").WithDescription("description2").WithAudit(now, u2, now, u2)

	testCases := []struct {
		name     string
		language string
		wantErr  bool
		want     []*models.Policy
	}{
		{name: "cedar", language: "cedar", want: []*models.Policy{wp1}},
		{name: "opa", language: "opa", want: []*models.Policy{wp2}},
		{name: "cerbos", language: "cerbos", want: []*models.Policy{}},
		{name: "all", want: []*models.Policy{wp1, wp2}},
		{name: "force error", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l := tc.language

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mock, db := newMockPG(t, ctx, "postgres://localhost:5432/myDB", time.Minute, 3, now)

			mock.ExpectBegin()

			var exp *pgxmock.ExpectedQuery
			sql := `SELECT language,id,title,description,rvva_id,uri,tags,content,created,created_by,updated,updated_by FROM policy`
			if tc.language != "" {
				sql += ` WHERE language=$1`
				exp = mock.ExpectQuery(sql).WithArgs(l)
			} else {
				exp = mock.ExpectQuery(sql)
			}

			if tc.wantErr {
				exp.WillReturnError(errors.New("test error"))
				mock.ExpectRollback()
			} else {
				rows := pgxmock.NewRows([]string{"language", "id", "title", "description", "rvva_id", "uri", "tags", "content", "created", "created_by", "updated", "updated_by"})
				for i := range tc.want {
					w := tc.want[i]
					rows.AddRow(w.Language(), w.ID(), w.Title(), w.Description(), w.RvvaID(), w.URI(), w.Tags(), w.ContentString(), w.Created(), w.CreatedBy(), w.Updated(), w.UpdatedBy())
				}
				exp.WillReturnRows(rows)
				mock.ExpectCommit()
			}

			got, err := db.ListPolicies(ctx, l)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.Len(t, got, len(tc.want))

				for i := range tc.want {
					policiesMatch(t, tc.want[i], got[i])
				}
			}
		})
	}
}

func newMockPG(t *testing.T, ctx context.Context, dsn string, maxLife time.Duration, maxConn int32, now time.Time) (pgxmock.PgxPoolIface, *PostgresDB) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	require.NotNil(t, mock)

	got, err2 := NewPostgresDB(ctx, dsn, maxLife, maxConn)
	require.NoError(t, err2)
	require.NotNil(t, got)

	got.p, err = postgresql.NewWithPool(ctx, dsn, maxLife, maxConn, mock)
	require.NoError(t, err)
	require.NotNil(t, got.p)

	got.now = func() time.Time { return now }
	return mock, got
}

func policiesMatch(t *testing.T, a, b *models.Policy) {
	require.NotNil(t, a)
	require.NotNil(t, b)

	assert.Equal(t, a.Language(), b.Language())
	assert.Equal(t, a.ID(), b.ID())
	assert.Equal(t, a.Title(), b.Title())
	assert.Equal(t, a.Description(), b.Description())
	assert.Equal(t, a.RvvaID(), b.RvvaID())
	assert.Equal(t, a.ContentString(), b.ContentString())
	assert.Equal(t, a.Created(), b.Created())
	assert.Equal(t, a.CreatedBy(), b.CreatedBy())
	assert.Equal(t, a.Updated(), b.Updated())
	assert.Equal(t, a.UpdatedBy(), b.UpdatedBy())
}
