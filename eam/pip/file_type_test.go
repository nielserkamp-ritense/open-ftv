package pip

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
)

func TestDetectType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		path     string
		wantType fileType
		wantMime string
		wantData bool
	}{
		{name: "empty text file", path: "../../testdata/unittest/bad/empty.txt", wantType: unknownFile},
		{name: "empty unknown file", path: "../../testdata/unittest/bad/empty.unknown.extension", wantType: unknownFile},
		{name: "bad yaml (1)", path: "../../testdata/unittest/pip/not_yaml.yaml", wantType: unknownFile},
		{name: "bad yaml (2)", path: "../../testdata/unittest/bad/not_yaml.yml", wantType: unknownFile},
		{name: "bad toml", path: "../../testdata/unittest/bad/not_toml.toml", wantType: unknownFile},
		{name: "bad json attributes", path: "../../testdata/unittest/pip2/misc/attr1.json", wantType: unknownFile},
		{name: "yaml attributes", path: "../../testdata/pip/attributes/werktijden.yaml", wantType: attributesFile, wantMime: mime.MimeTypeYAML, wantData: true},
		{name: "yaml entities (1)", path: "../../testdata/pip/entities/apps/app1.yaml", wantType: entitiesFile, wantMime: mime.MimeTypeYAML, wantData: true},
		{name: "yaml entities (2)", path: "../../testdata/pip/entities/services/brp.yaml", wantType: entitiesFile, wantMime: mime.MimeTypeYAML, wantData: true},
		{name: "toml attributes", path: "../../testdata/unittest/pip2/misc/attributes.toml", wantType: attributesFile, wantMime: mime.MimeTypeTOML, wantData: true},
		{name: "json attributes", path: "../../testdata/unittest/pip2/misc/attr2.json", wantType: attributesFile, wantMime: mime.MimeTypeJSON, wantData: true},
		{name: "openFGA model", path: "../../testdata/policies/openfga/doelbinding.model", wantType: policyFile, wantMime: mime.MimeTypeOpenFGA},
		{name: "openFGA relations", path: "../../testdata/policies/openfga/doelbinding.relations", wantType: relationsFile, wantMime: mime.MimeTypeJSON, wantData: true},
		{name: "cedar policy", path: "../../testdata/policies/cedar/brp/subsidies.cedar", wantType: policyFile, wantMime: mime.MimeTypeCedar},
		{name: "opa policy", path: "../../testdata/policies/opa/brp/subsidies.rego", wantType: policyFile, wantMime: mime.MimeTypeOPA},
		{name: "turtle collection (1)", path: "../../testdata/rdf/entities/services/rvig.ttl", wantType: collectionFile, wantMime: mime.MimeTypeTurtle},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ext := filepath.Ext(tc.path)
			mt := mime.ConvertExt(ext)

			f, err := os.Open(tc.path)
			require.NoError(t, err)
			defer f.Close()

			ft, mt2, data := detectType(f, mt)
			assert.Equal(t, tc.wantType, ft)
			assert.Equal(t, tc.wantMime, mt2)

			if tc.wantData {
				assert.NotNil(t, data)
			} else {
				assert.Nil(t, data)
			}
		})
	}
}

func TestDetectTypeFromJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		data     string
		wantErr  bool
		wantType fileType
		wantMime string
		wantData bool
	}{
		{
			name:    "empty",
			wantErr: true,
		},
		{
			name:    "bad JSON",
			data:    "/this/is/not/json",
			wantErr: true,
		},
		{
			name:    "unsupported type",
			data:    `true`,
			wantErr: true,
		},
		{
			name:    "unknown array type",
			data:    `[true]`,
			wantErr: true,
		},
		{
			name:     "good JSON",
			data:     `{"key":"werktijden","value":{"maandag":{"begin":"09:00","eind":"17:00"}}}`,
			wantType: attributesFile,
			wantMime: "application/json",
			wantData: true,
		},
		{
			name:     "good JSON-LD",
			data:     `{"@context":"production","value":"hello world"}`,
			wantType: collectionFile,
			wantMime: "application/ld+json",
			wantData: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ft, mt, d := detectTypeFromJSON(strings.NewReader(tc.data))
			if tc.wantErr {
				assert.Equal(t, unknownFile, ft)
				assert.Empty(t, mt)
				assert.Nil(t, d)
			} else {
				assert.Equal(t, tc.wantType, ft)
				assert.Equal(t, tc.wantMime, mt)

				if tc.wantData {
					assert.NotNil(t, d)
				} else {
					assert.Nil(t, d)
				}
			}
		})
	}
}

func TestDetectTypeFromYAML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		data     string
		wantErr  bool
		wantType fileType
		wantMime string
		wantData bool
	}{
		{
			name:    "empty",
			wantErr: true,
		},
		{
			name:    "bad YAML",
			data:    "\000\001\002",
			wantErr: true,
		},
		{
			name:    "unsupported type",
			data:    `true`,
			wantErr: true,
		},
		{
			name:    "unknown array type",
			data:    `[true]`,
			wantErr: true,
		},
		{
			name: "good YAML",
			data: `---
key: "werktijden"
value:
  - maandag:
    begin: "09:00"
    eind: "17:00"
`,
			wantType: attributesFile,
			wantMime: "application/yaml",
			wantData: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ft, mt, d := detectTypeFromYAML(strings.NewReader(tc.data))
			if tc.wantErr {
				assert.Equal(t, unknownFile, ft)
				assert.Empty(t, mt)
				assert.Nil(t, d)
			} else {
				assert.Equal(t, tc.wantType, ft)
				assert.Equal(t, tc.wantMime, mt)

				if tc.wantData {
					assert.NotNil(t, d)
				} else {
					assert.Nil(t, d)
				}
			}
		})
	}
}

func TestDetectTypeFromTOML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		data     string
		wantErr  bool
		wantType fileType
		wantMime string
		wantData bool
	}{
		{
			name:     "empty",
			wantType: attributesFile,
			wantMime: "application/toml",
		},
		{
			name:    "bad TOML",
			data:    "\000\001\002",
			wantErr: true,
		},
		{
			name:    "unsupported type",
			data:    `true`,
			wantErr: true,
		},
		{
			name: "good object",
			data: `key = "hello"
value = "world"`,
			wantType: attributesFile,
			wantMime: "application/toml",
			wantData: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ft, mt, d := detectTypeFromTOML(strings.NewReader(tc.data))
			if tc.wantErr {
				assert.Equal(t, unknownFile, ft)
				assert.Empty(t, mt)
				assert.Nil(t, d)
			} else {
				assert.Equal(t, tc.wantType, ft)
				assert.Equal(t, tc.wantMime, mt)

				if tc.wantData {
					assert.NotNil(t, d)
				} else {
					assert.Nil(t, d)
				}
			}
		})
	}
}
