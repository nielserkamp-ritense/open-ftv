package io

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSnifStream(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		path string
		want string
	}{
		{
			name: "bad file",
			path: "/what/have/we/here/dear/watson?",
			want: DefaultMimeType,
		},
		{
			name: "empty file",
			path: "../../testdata/unittest/bad/empty.txt",
			want: DefaultMimeType,
		},
		{
			name: "OPA/Rego - 1",
			path: "../../testdata/policies/opa/brp/subsidies.rego",
			want: MimeTypeOPA,
		},
		{
			name: "OPA/Rego - 2",
			path: "../../testdata/policies/opa/dienst_toeslagen/zorgtoeslag.rego",
			want: MimeTypeOPA,
		},
		{
			name: "Cedar - 1",
			path: "../../testdata/policies/cedar/brp/burgerzaken.cedar",
			want: MimeTypeCedar,
		},
		{
			name: "Cedar - 2",
			path: "../../testdata/policies/cedar/rdw/kentekens.cedar",
			want: MimeTypeCedar,
		},
		{
			name: "OpenFGA",
			path: "../../testdata/policies/openfga/doelbinding.model",
			want: MimeTypeOpenFGA,
		},
		{
			name: "bad toml",
			path: "../../testdata/unittest/bad/not_toml.toml",
			want: DefaultMimeType,
		},
		{
			name: "good toml",
			path: "../../testdata/unittest/pip2/misc/attributes.toml",
			want: MimeTypeTOML,
		},
		{
			name: "json - 1",
			path: "../../testdata/unittest/pip2/misc/attr1.json",
			want: MimeTypeJSON,
		},
		{
			name: "json - 2",
			path: "../../testdata/unittest/pip2/misc/attr2.json",
			want: MimeTypeJSON,
		},
		{
			name: "xml",
			path: "../../testdata/unittest/pip2/misc/attributes.xml",
			want: MimeTypeXML,
		},
		{
			name: "bad yaml",
			path: "../../testdata/unittest/bad/not_yaml.yml",
			want: DefaultMimeType,
		},
		{
			name: "good yaml",
			path: "../../testdata/unittest/pip2/entities/activities.yaml",
			want: MimeTypeYAML,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f, err := os.Open(tc.path)
			if err == nil {
				defer f.Close()
			}

			got := SnifStream(f)
			assert.Equal(t, tc.want, got)
		})
	}
}
