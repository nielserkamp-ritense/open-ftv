package pap

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/policies"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNew(t *testing.T) {
	t.Run("new PAP", func(t *testing.T) {
		h := slog2.NewDummyHandler(0)

		p := New(nil, slog.New(h), nil)
		require.NotNil(t, p)

		p2, ok := p.(*pap)
		require.True(t, ok)
		require.NotNil(t, p2)

		assert.NotNil(t, p2.logger)
		assert.Nil(t, p2.events)
		assert.NotNil(t, p2.policies)
	})
}

func TestPap_Add(t *testing.T) {
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
			h := slog2.NewDummyHandler(0)
			e := &eventCounter{}

			p := New(nil, slog.New(h), e)
			require.NotNil(t, p)

			pol, err2 := NewPolicy(&policies.Policy{Id: tc.id}, tc.data)
			if tc.wantErr {
				require.Error(t, err2)
				require.Nil(t, pol)
			} else {
				require.NoError(t, err2)
				require.NotNil(t, pol)

				pol2, err3 := p.Add(pol)
				require.NoError(t, err3)
				require.NotNil(t, pol2)

				assert.Equal(t, tc.wantCount, e.added)
				assert.Zero(t, e.replaced)
				assert.Zero(t, e.removed)

				p2, ok := p.(*pap)
				require.True(t, ok)
				require.NotNil(t, p2)

				assert.Equal(t, tc.wantCount, len(p2.policies))
			}
		})
	}
}

func TestPap_Replace(t *testing.T) {
	closedFile, err := os.Open("/etc/hostname")
	require.NoError(t, err)
	closedFile.Close()

	testCases := []struct {
		name      string
		cached    []string
		id        string
		data      io.Reader
		wantErr   bool
		wantCount int
		want      string
	}{
		{
			name:    "empty cache",
			id:      "x1.txt",
			data:    bytes.NewReader([]byte("some data")),
			wantErr: true,
		},
		{
			name:    "key missing",
			cached:  []string{"x2.txt", "x3.txt"},
			id:      "x1.txt",
			data:    bytes.NewReader([]byte("some data")),
			wantErr: true,
		},
		{
			name:      "key found",
			cached:    []string{"x2.txt", "x3.txt", "x1.txt"},
			id:        "x1.txt",
			data:      bytes.NewReader([]byte("some data")),
			wantCount: 3,
			want:      "some data",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(0)
			e := &eventCounter{}

			p := New(nil, slog.New(h), e)
			require.NotNil(t, p)

			for i := range tc.cached {
				id := tc.cached[i]

				pol, err2 := NewPolicy(&policies.Policy{Id: id}, bytes.NewBuffer([]byte("data")))
				require.NoError(t, err2)
				require.NotNil(t, pol)

				_, err2 = p.Add(pol)
				require.NoError(t, err2)
			}

			pol, err2 := NewPolicy(&policies.Policy{Id: tc.id}, tc.data)
			require.NoError(t, err2)
			require.NotNil(t, pol)

			pol2, err3 := p.Replace(pol)
			if tc.wantErr {
				require.Error(t, err3)
				require.Nil(t, pol2)
			} else {
				require.NoError(t, err3)
				require.NotNil(t, pol2)

				assert.Equal(t, len(tc.cached), e.added)
				assert.Equal(t, 1, e.replaced)
				assert.Zero(t, e.removed)

				p2, ok := p.(*pap)
				require.True(t, ok)
				require.NotNil(t, p2)

				assert.Equal(t, tc.wantCount, len(p2.policies))

				f, err4 := p.Get(tc.id)
				require.NoError(t, err4)
				require.NotNil(t, f)

				data, err5 := io.ReadAll(f.Content())
				require.NoError(t, err5)
				require.Equal(t, tc.want, string(data))
			}
		})
	}
}

func TestPap_Remove(t *testing.T) {
	testCases := []struct {
		name      string
		cached    []string
		key       string
		wantErr   bool
		wantCount int
	}{
		{name: "empty cache", key: "x1.txt", wantErr: true},
		{name: "key missing", cached: []string{"x2.txt", "x3.txt"}, key: "x1.txt", wantErr: true},
		{name: "key found", cached: []string{"x2.txt", "x3.txt", "x1.txt"}, key: "x1.txt", wantCount: 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(0)
			e := &eventCounter{}

			p := New(nil, slog.New(h), e)
			require.NotNil(t, p)

			for i := range tc.cached {
				id := tc.cached[i]

				pol, err2 := NewPolicy(&policies.Policy{Id: id}, bytes.NewBuffer([]byte("data")))
				require.NoError(t, err2)
				require.NotNil(t, pol)

				_, err2 = p.Add(pol)
				require.NoError(t, err2)
			}

			pol, err2 := p.Remove(tc.key)
			if tc.wantErr {
				require.Error(t, err2)
				require.Nil(t, pol)
			} else {
				require.NoError(t, err2)
				require.NotNil(t, pol)

				assert.Equal(t, len(tc.cached), e.added)
				assert.Zero(t, e.replaced)
				assert.Equal(t, 1, e.removed)

				p2, ok := p.(*pap)
				require.True(t, ok)
				require.NotNil(t, p2)

				assert.Equal(t, tc.wantCount, len(p2.policies))

				f, err4 := p.Get(tc.key)
				require.Error(t, err4)
				require.Nil(t, f)
			}
		})
	}
}

func TestPap_ListAllKeys(t *testing.T) {
	testCases := []struct {
		name   string
		cached []string
		want   []string
	}{
		{name: "empty", want: []string{}},
		{name: "one", cached: []string{"x2.txt"}, want: []string{"x2.txt"}},
		{name: "few", cached: []string{"x2.txt", "x3.txt", "x1.txt"}, want: []string{"x1.txt", "x2.txt", "x3.txt"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(0)

			p := New(nil, slog.New(h), nil)
			require.NotNil(t, p)

			for i := range tc.cached {
				id := tc.cached[i]

				pol, err2 := NewPolicy(&policies.Policy{Id: id}, bytes.NewBuffer([]byte("data")))
				require.NoError(t, err2)
				require.NotNil(t, pol)

				_, err2 = p.Add(pol)
				require.NoError(t, err2)
			}

			got := p.ListAllKeys()
			assert.EqualValues(t, tc.want, got)
		})
	}
}

type eventCounter struct {
	added    int
	replaced int
	removed  int
}

func (e *eventCounter) Handle(t models.EventType, _ string) {
	switch t {
	case models.PolicyAdded:
		e.added++
	case models.PolicyReplaced:
		e.replaced++
	case models.PolicyRemoved:
		e.removed++
	default:
	}
}
