package pap

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("new PAP", func(t *testing.T) {
		h := slog2.NewDummyHandler(0)

		p := New(nil, slog.New(h))
		require.NotNil(t, p)

		assert.NotNil(t, p.logger)
		assert.NotNil(t, p.eventSinks)
		assert.NotNil(t, p.updates)
		assert.NotNil(t, p.deletes)
	})
}

func TestPap_Add(t *testing.T) {
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

			h := slog2.NewDummyHandler(0)
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

				pol2, err3 := p.Create(pol)
				require.NoError(t, err3)
				require.NotNil(t, pol2)

				assert.Equal(t, tc.wantCount, e.created)
				assert.Zero(t, e.updated)
				assert.Zero(t, e.deleted)
			}
		})
	}
}

func TestPap_Replace(t *testing.T) {
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

			h := slog2.NewDummyHandler(0)
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

				_, err2 = p.Create(pol)
				require.NoError(t, err2)
			}

			parts := strings.Split(tc.key, "/")

			prev, err2 := models.NewPolicyFromOAS(&policies.Policy{Language: parts[0], Id: parts[1]}, bytes.NewBuffer([]byte("data")))
			require.NoError(t, err2)
			require.NotNil(t, prev)

			pol, err3 := models.NewPolicyFromOAS(&policies.Policy{Language: parts[0], Id: parts[1]}, tc.data)
			require.NoError(t, err3)
			require.NotNil(t, pol)

			pol2, err4 := p.Update(prev, 0, pol)
			if tc.wantErr {
				require.Error(t, err4)
				require.Nil(t, pol2)
			} else {
				require.NoError(t, err4)
				require.NotNil(t, pol2)

				assert.Equal(t, len(tc.cached), e.created)
				assert.Equal(t, 1, e.updated)
				assert.Zero(t, e.deleted)

				f, _, err5 := p.Read(parts[0], parts[1])
				require.NoError(t, err5)
				require.NotNil(t, f)

				data, err6 := io.ReadAll(f.Content())
				require.NoError(t, err6)
				require.Equal(t, tc.want, string(data))
			}
		})
	}
}

func TestPap_Remove(t *testing.T) {
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

			h := slog2.NewDummyHandler(0)
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

				_, err2 = p.Create(pol)
				require.NoError(t, err2)
			}

			parts := strings.Split(tc.key, "/")
			prev, err2 := models.NewPolicyFromOAS(&policies.Policy{Language: parts[0], Id: parts[1]}, bytes.NewBuffer([]byte("data")))
			require.NoError(t, err2)
			require.NotNil(t, prev)

			pol, err3 := p.Delete(prev, 0)
			if tc.wantErr {
				require.Error(t, err3)
				require.Nil(t, pol)
			} else {
				require.NoError(t, err3)
				require.NotNil(t, pol)

				assert.Equal(t, len(tc.cached), e.created)
				assert.Zero(t, e.updated)
				assert.Equal(t, 1, e.deleted)

				f, _, err4 := p.Read("", tc.key)
				require.Error(t, err4)
				require.Nil(t, f)
			}
		})
	}
}

func TestPap_ListAllKeys(t *testing.T) {
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

			h := slog2.NewDummyHandler(0)

			p := New(nil, slog.New(h))
			require.NotNil(t, p)

			for i := range tc.cached {
				key := tc.cached[i]
				parts := strings.Split(key, "/")

				pol, err2 := models.NewPolicyFromOAS(&policies.Policy{Language: parts[0], Id: parts[1]}, bytes.NewBuffer([]byte("data")))
				require.NoError(t, err2)
				require.NotNil(t, pol)

				_, err2 = p.Create(pol)
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

func TestPAP_NewDeployment(t *testing.T) {
	t.Parallel()

	t.Run("new deployment", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(0)
		logger := slog.New(h)

		p := New(nil, logger)
		require.NotNil(t, p)

		m := bundles.NewManager(ctx, logger, bundles.WithConfig("../../testdata/unittests/bundles/test1", false))
		require.NotNil(t, m)

		d, err := p.NewDeployment("merry easter", m)
		require.NoError(t, err)
		require.NotNil(t, d)

		assert.Equal(t, uint64(1), d.Version())
		assert.Equal(t, bundles.Creating, d.Status())
	})
}

func TestPAP_LastDeployment(t *testing.T) {
	t.Parallel()

	t.Run("last deployment", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(0)
		logger := slog.New(h)

		p := New(nil, logger)
		require.NotNil(t, p)

		m := bundles.NewManager(ctx, logger, bundles.WithConfig("../../testdata/unittests/bundles/test1", false))
		require.NotNil(t, m)

		d, err := p.NewDeployment("merry easter", m)
		require.NoError(t, err)
		require.NotNil(t, d)

		d2, err2 := p.LastDeployment()
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, uint64(1), d2.Version())
		assert.GreaterOrEqual(t, d2.Status(), bundles.Creating)
	})
}

func TestPAP_ReadDeployment(t *testing.T) {
	t.Parallel()

	t.Run("read deployment", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(0)
		logger := slog.New(h)

		p := New(nil, logger)
		require.NotNil(t, p)

		m := bundles.NewManager(ctx, logger, bundles.WithConfig("../../testdata/unittests/bundles/test1", false))
		require.NotNil(t, m)

		d, err := p.NewDeployment("merry easter", m)
		require.NoError(t, err)
		require.NotNil(t, d)

		d2, err2 := p.ReadDeployment(1)
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, uint64(1), d2.Version())
		assert.Equal(t, bundles.Creating, d2.Status())

		d3, err3 := p.ReadDeployment(2)
		require.Error(t, err3)
		require.Nil(t, d3)
	})
}

func TestPAP_ListDeployments(t *testing.T) {
	t.Parallel()

	t.Run("list deployments", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(0)
		logger := slog.New(h)

		p := New(nil, logger)
		require.NotNil(t, p)

		m := bundles.NewManager(ctx, logger, bundles.WithConfig("../../testdata/unittests/bundles/test1", false))
		require.NotNil(t, m)

		d, err := p.NewDeployment("merry easter", m)
		require.NoError(t, err)
		require.NotNil(t, d)

		list, err2 := p.ListDeployments()
		require.NoError(t, err2)
		require.NotNil(t, list)
		assert.Equal(t, 1, len(list))
	})
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

			h := slog2.NewDummyHandler(0)
			e := &eventCounter{}

			p := New(nil, slog.New(h))
			require.NotNil(t, p)

			p.AddEventSink(e)

			_, err := p.Create(p1)
			require.NoError(t, err)
			_, err = p.Create(p2)
			require.NoError(t, err)
			_, err = p.Create(p3)
			require.NoError(t, err)

			assert.Equal(t, 3, e.created)
			assert.Zero(t, e.updated)
			assert.Zero(t, e.deleted)

			e.created = 0

			err = p.ReplaceAll(tc.list)
			require.NoError(t, err)

			assert.Equal(t, tc.want, e.created)
			assert.Zero(t, e.updated)
			assert.Equal(t, 3, e.deleted)
		})
	}
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
