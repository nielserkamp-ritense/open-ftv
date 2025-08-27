package pap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestPAP_Add(t *testing.T) {
	t.Parallel()

	closedFile, err := os.Open("/etc/hostname")
	require.NoError(t, err)
	closedFile.Close()

	testCases := []struct {
		name      string
		id        string
		data      io.Reader
		wantErr   bool
		wantCount int
	}{
		{name: "nil", id: "x1.txt", wantErr: true},
		{name: "closed file", id: "x2.txt", data: closedFile, wantErr: true},
		{name: "good file", id: "x3.txt", data: bytes.NewBuffer([]byte("some data")), wantCount: 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)
			e := &eventCounter{}

			p := New(nil, slog.New(h))
			require.NotNil(t, p)

			p.AddEventSink(e)

			pol, err2 := models.NewPolicyFromOAS(&policies.Policy{Id: tc.id}, tc.data)
			if tc.wantErr {
				require.Error(t, err2)
				require.Nil(t, pol)
			} else {
				require.NoError(t, err2)
				require.NotNil(t, pol)

				pol2, err3 := p.Create(pol, "test")
				require.NoError(t, err3)
				require.NotNil(t, pol2)

				assert.Equal(t, tc.wantCount, e.created)
				assert.Zero(t, e.updated)
				assert.Zero(t, e.deleted)
			}
		})
	}
}

func TestPAP_Replace(t *testing.T) {
	t.Parallel()

	closedFile, err := os.Open("/etc/hostname")
	require.NoError(t, err)
	closedFile.Close()

	testCases := []struct {
		name      string
		cached    []string
		key       string
		data      io.Reader
		wantErr   bool
		wantCount int
		want      string
	}{
		{
			name:    "empty cache",
			key:     "rego/x1.txt",
			data:    bytes.NewReader([]byte("some data")),
			wantErr: true,
		},
		{
			name:    "key missing",
			cached:  []string{"rego/x2.txt", "rego/x3.txt"},
			key:     "rego/x1.txt",
			data:    bytes.NewReader([]byte("some data")),
			wantErr: true,
		},
		{
			name:      "key found",
			cached:    []string{"rego/x2.txt", "rego/x3.txt", "rego/x1.txt"},
			key:       "rego/x1.txt",
			data:      bytes.NewReader([]byte("some data")),
			wantCount: 3,
			want:      "some data",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)
			e := &eventCounter{}

			p := New(nil, slog.New(h))
			require.NotNil(t, p)

			p.AddEventSink(e)

			for i := range tc.cached {
				key := tc.cached[i]
				parts := strings.Split(key, "/")

				pol, err2 := models.NewPolicyFromOAS(&policies.Policy{Language: parts[0], Id: parts[1]}, bytes.NewBuffer([]byte("data")))
				require.NoError(t, err2)
				require.NotNil(t, pol)

				_, err2 = p.Create(pol, "test")
				require.NoError(t, err2)
			}

			parts := strings.Split(tc.key, "/")

			prev, err2 := models.NewPolicyFromOAS(&policies.Policy{Language: parts[0], Id: parts[1]}, bytes.NewBuffer([]byte("data")))
			require.NoError(t, err2)
			require.NotNil(t, prev)

			pol, err3 := models.NewPolicyFromOAS(&policies.Policy{Language: parts[0], Id: parts[1]}, tc.data)
			require.NoError(t, err3)
			require.NotNil(t, pol)

			pol2, err4 := p.Update(prev, 0, pol, "test")
			if tc.wantErr {
				require.Error(t, err4)
				require.Nil(t, pol2)
			} else {
				require.NoError(t, err4)
				require.NotNil(t, pol2)

				assert.Equal(t, len(tc.cached), e.created)
				assert.Equal(t, 1, e.updated)
				assert.Zero(t, e.deleted)

				f, _, err5 := p.Read(parts[1])
				require.NoError(t, err5)
				require.NotNil(t, f)

				data, err6 := io.ReadAll(f.Content())
				require.NoError(t, err6)
				require.Equal(t, tc.want, string(data))
			}
		})
	}
}

func TestPAP_Remove(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		cached    []string
		key       string
		wantErr   bool
		wantCount int
	}{
		{name: "empty cache", key: "rego/x1.txt", wantErr: true},
		{name: "key missing", cached: []string{"rego/x2.txt", "rego/x3.txt"}, key: "rego/x1.txt", wantErr: true},
		{name: "key found", cached: []string{"rego/x2.txt", "rego/x3.txt", "rego/x1.txt"}, key: "rego/x1.txt", wantCount: 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)
			e := &eventCounter{}

			p := New(nil, slog.New(h))
			require.NotNil(t, p)

			p.AddEventSink(e)

			for i := range tc.cached {
				key := tc.cached[i]
				parts := strings.Split(key, "/")

				pol, err2 := models.NewPolicyFromOAS(&policies.Policy{Language: parts[0], Id: parts[1]}, bytes.NewBuffer([]byte("data")))
				require.NoError(t, err2)
				require.NotNil(t, pol)

				_, err2 = p.Create(pol, "test")
				require.NoError(t, err2)
			}

			parts := strings.Split(tc.key, "/")
			prev, err2 := models.NewPolicyFromOAS(&policies.Policy{Language: parts[0], Id: parts[1]}, bytes.NewBuffer([]byte("data")))
			require.NoError(t, err2)
			require.NotNil(t, prev)

			pol, err3 := p.Delete(prev, 0, "test")
			if tc.wantErr {
				require.Error(t, err3)
				require.Nil(t, pol)
			} else {
				require.NoError(t, err3)
				require.NotNil(t, pol)

				assert.Equal(t, len(tc.cached), e.created)
				assert.Zero(t, e.updated)
				assert.Equal(t, 1, e.deleted)

				f, _, err4 := p.Read(tc.key)
				require.NoError(t, err4)
				require.Nil(t, f)
			}
		})
	}
}

func TestPAP_List(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		cached []string
		want   map[string]bool
	}{
		{name: "empty", want: map[string]bool{}},
		{name: "one", cached: []string{"rego/x2.txt"}, want: map[string]bool{"rego/x2.txt": true}},
		{name: "few", cached: []string{"rego/x2.txt", "rego/x3.txt", "rego/x1.txt"}, want: map[string]bool{"rego/x1.txt": true, "rego/x2.txt": true, "rego/x3.txt": true}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)

			p := New(nil, slog.New(h))
			require.NotNil(t, p)

			for i := range tc.cached {
				key := tc.cached[i]
				parts := strings.Split(key, "/")

				pol, err2 := models.NewPolicyFromOAS(&policies.Policy{Language: parts[0], Id: parts[1]}, bytes.NewBuffer([]byte("data")))
				require.NoError(t, err2)
				require.NotNil(t, pol)

				_, err2 = p.Create(pol, "test")
				require.NoError(t, err2)
			}

			got, err := p.List("")
			require.NoError(t, err)
			assert.Equal(t, len(tc.want), len(got))

			for i := range got {
				pol := got[i]

				_, ok := tc.want[pol.Key()]
				assert.True(t, ok)
			}
		})
	}
}

func TestPAP_Iterate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		cached []string
		want   map[string]bool
	}{
		{name: "empty", want: map[string]bool{}},
		{name: "one", cached: []string{"rego/x2.txt"}, want: map[string]bool{"rego/x2.txt": true}},
		{name: "few", cached: []string{"rego/x2.txt", "rego/x3.txt", "rego/x1.txt"}, want: map[string]bool{"rego/x1.txt": true, "rego/x2.txt": true, "rego/x3.txt": true}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)

			p := New(nil, slog.New(h))
			require.NotNil(t, p)

			for i := range tc.cached {
				key := tc.cached[i]
				parts := strings.Split(key, "/")

				pol, err2 := models.NewPolicyFromOAS(&policies.Policy{Language: parts[0], Id: parts[1]}, bytes.NewBuffer([]byte("data")))
				require.NoError(t, err2)
				require.NotNil(t, pol)

				_, err2 = p.Create(pol, "test")
				require.NoError(t, err2)
			}

			var count int
			p.Iterate(func(p *models.Policy) {
				require.NotNil(t, p)
				count++
			})
			assert.Equal(t, len(tc.want), count)
		})
	}
}

func TestPAP_ReplaceAll(t *testing.T) {
	t.Parallel()

	p1, err1 := models.NewPolicyFromData("p1", "cedar", "", "", bytes.NewBufferString("allow = true;"))
	require.NoError(t, err1)
	require.NotNil(t, p1)

	p2, err2 := models.NewPolicyFromData("p2", "cedar", "", "", bytes.NewBufferString("allow = true;"))
	require.NoError(t, err2)
	require.NotNil(t, p2)

	p3, err3 := models.NewPolicyFromData("p3", "cedar", "", "", bytes.NewBufferString("allow = true;"))
	require.NoError(t, err3)
	require.NotNil(t, p3)

	p4, err4 := models.NewPolicyFromData("p4", "cedar", "", "", bytes.NewBufferString("allow = true;"))
	require.NoError(t, err4)
	require.NotNil(t, p4)

	testCases := []struct {
		name string
		list []*models.Policy
		want int
	}{
		{name: "nil", want: 0},
		{name: "empty", list: []*models.Policy{}},
		{name: "one", list: []*models.Policy{p4}, want: 1},
		{name: "three", list: []*models.Policy{p3, p4, p1}, want: 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)
			e := &eventCounter{}

			p := New(nil, slog.New(h))
			require.NotNil(t, p)

			p.AddEventSink(e)

			_, err := p.Create(p1, "test")
			require.NoError(t, err)
			_, err = p.Create(p2, "test")
			require.NoError(t, err)
			_, err = p.Create(p3, "test")
			require.NoError(t, err)

			assert.Equal(t, 3, e.created)
			assert.Zero(t, e.updated)
			assert.Zero(t, e.deleted)

			e.created = 0

			err = p.ReplaceAll(tc.list, "test")
			require.NoError(t, err)

			assert.Equal(t, tc.want, e.created)
			assert.Zero(t, e.updated)
			assert.Equal(t, 3, e.deleted)
		})
	}
}

func TestPAP_WithPostgresDB(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	u1 := "donald@duck.us"
	u2 := "goofy@weird.us"

	p1, e1 := models.NewPolicyFromData("p1", "cedar", "rvva1", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, e1)
	p2, e2 := models.NewPolicyFromData("p1", "cedar", "rvva2", "", bytes.NewBufferString("allow=false;"))
	require.NoError(t, e2)

	p1.WithAudit(now, u2, now, u2)
	p2.WithTags("x", "y", "z").WithTitle("title2").WithDescription("description2").WithAudit(now, u1, now, u1)

	wp1, we1 := models.NewPolicyFromData("p1", "cedar", "rvva1", "", bytes.NewBufferString("allow=true;"))
	require.NoError(t, we1)
	wp2, we2 := models.NewPolicyFromData("p1", "cedar", "rvva2", "", bytes.NewBufferString("allow=false;"))
	require.NoError(t, we2)

	wp1.WithAudit(now, u1, now, u1)
	wp2.WithTags("x", "y", "z").WithTitle("title2").WithDescription("description2").WithAudit(now, u1, now, u2)

	t.Run("with postgres DB", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelInfo)
		e := &eventCounter{}
		logger := slog.New(h)

		mock, db := newMockPG(t, ctx, "postgres://localhost:5432/myDB", time.Minute, 3, now)

		p := New(ctx, logger, WithPolicyDB(db))
		require.NotNil(t, p)

		p.AddEventSink(e)

		// create
		mock.ExpectBegin()
		mock.ExpectExec(fmt.Sprintf("SELECT set_config('openftv.user', '%s', true);", u1)).WillReturnResult(pgxmock.NewResult("SELECT", 1))
		mock.ExpectExec(`INSERT INTO policy
 (language,id,title,description,rvva_id,uri,tags,content,created,created_by,updated,updated_by)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`).
			WithArgs(p1.Language(), p1.ID(), p1.Title(), p1.Description(), p1.RvvaID(), p1.URI(), p1.Tags(), p1.ContentString(), now, u1, now, u1).
			WillReturnResult(pgxmock.NewResult("CREATE", 1))
		mock.ExpectCommit()

		// read
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT language,id,title,description,rvva_id,uri,tags,content,created,created_by,updated,updated_by
 FROM policy
 WHERE id=$1`).WithArgs("p1").
			WillReturnRows(
				pgxmock.NewRows([]string{"language", "id", "title", "description", "rvvaID", "uri", "tags", "content", "created", "createdBy", "updated", "updatedBy"}).
					AddRow(p1.Language(), p1.ID(), p1.Title(), p1.Description(), p1.RvvaID(), p1.URI(), p1.Tags(), p1.ContentString(), now, u1, now, u1))
		mock.ExpectCommit()

		// update
		mock.ExpectBegin()
		mock.ExpectExec(fmt.Sprintf("SELECT set_config('openftv.user', '%s', true);", u2)).WillReturnResult(pgxmock.NewResult("SELECT", 1))
		mock.ExpectExec(`UPDATE policy
 SET language=$3,title=$4,description=$5,rvva_id=$6,uri=$7,tags=$8,content=$9,updated=$10,updated_by=$11
 WHERE id=$1 AND updated=$2`).
			WithArgs(p2.ID(), now, p2.Language(), p2.Title(), p2.Description(), p2.RvvaID(), p2.URI(), p2.Tags(), p2.ContentString(), now, u2).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		mock.ExpectCommit()

		mock.ExpectBegin()
		mock.ExpectExec(fmt.Sprintf("SELECT set_config('openftv.user', '%s', true);", u1)).WillReturnResult(pgxmock.NewResult("SELECT", 1))
		mock.ExpectExec(`DELETE policy
 WHERE id=$1 AND updated=$2`).WithArgs(p2.ID(), now).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		mock.ExpectCommit()

		got, err := p.Create(p1, u1)
		require.NoError(t, err)
		require.NotNil(t, got)
		policiesMatch(t, wp1, got)

		got, _, err = p.Read("p1")
		require.NoError(t, err)
		require.NotNil(t, got)
		policiesMatch(t, wp1, got)

		got, err = p.Update(p1, timeToLastIndex(now), p2, u2)
		require.NoError(t, err)
		require.NotNil(t, got)
		policiesMatch(t, wp2, got)

		got, err = p.Delete(wp2, timeToLastIndex(now), u1)
		require.NoError(t, err)
		require.NotNil(t, got)
		policiesMatch(t, wp2, got)
	})
}

type eventCounter struct {
	created int
	updated int
	deleted int
}

func (e *eventCounter) Handle(t models.EventType, _ string) {
	switch t {
	case models.PolicyAdded:
		e.created++
	case models.PolicyReplaced:
		e.updated++
	case models.PolicyRemoved:
		e.deleted++
	default:
	}
}
