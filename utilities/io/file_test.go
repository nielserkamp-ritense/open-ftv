package io

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSnifFile(t *testing.T) {
	testCases := []struct {
		name string
		path string
		want string
	}{
		{
			name: "bad path",
			path: "/this/is/not/a/real/file/duh",
			// file open error!
			want: DefaultMimeType,
		},
		{
			name: "empty unknown file",
			path: "../../testdata/unittest/bad/empty.unknown.extension",
			// based on content!
			want: DefaultMimeType,
		},
		{
			name: "empty txt file",
			path: "../../testdata/unittest/bad/empty.txt",
			// based on extension!
			want: MimeTypePlain,
		},
		{
			name: "bad toml",
			path: "../../testdata/unittest/bad/not_toml.toml",
			// based on extension!
			want: MimeTypeTOML,
		},
		{
			name: "bad yaml",
			path: "../../testdata/unittest/bad/not_yaml.yml",
			// based on extension!
			want: MimeTypeYAML,
		},
		{
			name: "OpenFGA",
			path: "../../testdata/unittest/openfga/doelbinding.model",
			// based on content!
			want: MimeTypeOpenFGA,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := SnifFile(tc.path)
			assert.Equal(t, tc.want, got)
		})
	}
}
