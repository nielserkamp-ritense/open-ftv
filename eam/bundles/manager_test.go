package bundles

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNewManager(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		path    string
		recurse bool
		want    []string
		wantLog int
	}{
		{
			name:    "bad path",
			path:    "/not/a/valid/directory/path",
			want:    []string{},
			wantLog: 1,
		},
		{
			name:    "bad yaml",
			path:    "../../testdata/unittest/bundles/test3",
			want:    []string{},
			wantLog: 1,
		},
		{
			name:    "bad json",
			path:    "../../testdata/unittest/bundles/test4",
			want:    []string{},
			wantLog: 1,
		},
		{
			name:    "single",
			path:    "../../testdata/unittest/bundles/test1",
			recurse: true,
			want:    []string{"brp"},
		},
		{
			name: "few - no recurse",
			path: "../../testdata/unittest/bundles/test2",
			want: []string{"brp", "rdw"},
		},
		{
			name:    "few - recurse",
			path:    "../../testdata/unittest/bundles/test2",
			recurse: true,
			want:    []string{"brp", "rdw", "ui"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			m := NewManager(ctx, logger, WithConfig(tc.path, tc.recurse))
			require.NotNil(t, m)
			assert.GreaterOrEqual(t, h.Count(), tc.wantLog)

			for i := range tc.want {
				b := m.bundles[tc.want[i]]
				assert.NotNilf(t, b, tc.want[i])
			}

			list := m.Bundles()
			require.Len(t, list, len(tc.want))

			for i := range list {
				assert.Equal(t, tc.want[i], list[i].ID)
			}
		})
	}
}

func TestManager_LoadYAML_Fail(t *testing.T) {
	t.Parallel()

	t.Run("load yaml fail", func(t *testing.T) {
		t.Parallel()

		m := &Manager{}
		err := m.loadYAML("/not/a/valid/file/path.yaml")
		require.Error(t, err)
	})
}

func TestManager_LoadJSON_Fail(t *testing.T) {
	t.Parallel()

	t.Run("load json fail", func(t *testing.T) {
		t.Parallel()

		m := &Manager{}
		err := m.loadJSON("/not/a/valid/file/path.json")
		require.Error(t, err)
	})
}
