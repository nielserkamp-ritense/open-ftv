package pap

import (
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
