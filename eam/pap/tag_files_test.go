package pap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadTagFiles(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		path    string
		wantErr bool
		want    int
	}{
		{
			name: "no path",
		},
		{
			name:    "missing path",
			path:    "/this/is/not/a/valid/path/duh",
			wantErr: true,
		},
		{
			name:    "bad yaml",
			path:    "../../testdata/unittest/bad/tags/yaml",
			wantErr: true,
		},
		{
			name:    "bad json",
			path:    "../../testdata/unittest/bad/tags/json",
			wantErr: true,
		},
		{
			name: "good",
			path: "../../testdata/apps/pap/tags",
			want: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tags, err := ReadTagFiles(tc.path)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(tags), tc.want)
			}
		})
	}
}

func TestReadTagFilesConfigMapMount(t *testing.T) {
	t.Parallel()

	// The layout Kubernetes gives a mounted ConfigMap.
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "..2026_09_30_12_00_00.123456789"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "..2026_09_30_12_00_00.123456789", "tags.yaml"), []byte("- id: ux\n  name: user interface\n"), 0o600))
	require.NoError(t, os.Symlink("..2026_09_30_12_00_00.123456789", filepath.Join(dir, "..data")))
	require.NoError(t, os.Symlink(filepath.Join("..data", "tags.yaml"), filepath.Join(dir, "tags.yaml")))

	tags, err := ReadTagFiles(dir)
	require.NoError(t, err)
	assert.Len(t, tags, 1)
}
