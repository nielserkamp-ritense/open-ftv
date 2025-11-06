package schema

import (
	"bytes"
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
)

func TestEndpoint_Fix(t *testing.T) {
	t.Parallel()

	f1 := &Field{Object: Object{Parent: Parent{ID: "f1"}, Description: "field 1"}, Type: enums.IntegerType}
	f2 := &Field{Object: Object{Parent: Parent{ID: "f2"}, Description: "field 2"}, Type: enums.StringType}
	f3 := &Field{Object: Object{Parent: Parent{ID: "f1"}, Description: "field 1"}, Type: enums.IntegerType}
	f4 := &Field{Object: Object{Parent: Parent{ID: "f2"}, Description: "field 2"}, Type: enums.StringType}

	t1 := &Table{Object: Object{Parent: Parent{ID: "t1"}, Description: "table 1", Fields: []*Field{f1, f2}}}
	t2 := &Table{Object: Object{Parent: Parent{ID: "t2"}, Description: "table 2", Fields: []*Field{f3, f4}}}

	ds1 := &Datasource{Parent: Parent{ID: "src1"}, Description: "source 1", Tables: []*Table{t1, t2}}

	testCases := []struct {
		name         string
		endpoint     *Endpoint
		wantUID      string
		wantPrimary  *Table
		wantIncludes []*Field
		wantExcludes []*Field
	}{
		{
			name: "short table update",
			endpoint: &Endpoint{
				Version:     1,
				Type:        enums.PutMethod,
				Path:        "/",
				FullVersion: "1.0.0",
				Description: "shortened table 1",
				Datasource:  "src1",
				Table:       "t1",
				Fields:      []string{"t1.f1"},
			},
			wantUID:      "PUT:/v1/",
			wantPrimary:  t1,
			wantIncludes: []*Field{f1},
		},
		{
			name: "single table with filter",
			endpoint: &Endpoint{
				Version:     1,
				Type:        enums.GetMethod,
				CalledAs:    enums.PostMethod,
				Path:        "yoyo/",
				FullVersion: "1.0.0",
				Description: "filtered table 2",
				Datasource:  "src1",
				Table:       "t2",
				Fields:      []string{"f1", "f2"},
				Filter:      map[string]any{"f1": "1"},
			},
			wantUID:      "POST:/v1/yoyo",
			wantPrimary:  t2,
			wantIncludes: []*Field{f3, f4},
		},
		{
			name: "simple short sibling join",
			endpoint: &Endpoint{
				Version:     1,
				Path:        "/yoyo",
				FullVersion: "1.0.0",
				Description: "join table 1 and 2",
				Datasource:  "src1",
				Table:       "t1",
				Joins:       []*Join{{Type: enums.ForcedSibling, QualifiedFields: true, Target: "t1", Source: "t2"}},
				Fields:      []string{"*", "!t2.f1"},
			},
			wantUID:      "GET:/v1/yoyo",
			wantPrimary:  t1,
			wantIncludes: []*Field{f1, f2, f3, f4},
			wantExcludes: []*Field{f3},
		},
		{
			name: "simple full parent/child join",
			endpoint: &Endpoint{
				Version:     1,
				Path:        "/1/2/3/",
				FullVersion: "1.0.0",
				Description: "join table 1 and 2",
				Datasource:  "src1",
				Table:       "t1",
				Joins:       []*Join{{Type: enums.ForcedParentChild, Target: "t1", Source: "t2"}},
			},
			wantUID:      "GET:/v1/1/2/3",
			wantPrimary:  t1,
			wantIncludes: []*Field{f1, f2, f3, f4},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := tc.endpoint
			e.Fix(ds1)

			assert.Equal(t, ds1, e.datasource)
			assert.Equal(t, tc.wantUID, e.UID())
			assert.EqualValues(t, tc.wantPrimary, e.Primary())

			require.Equal(t, len(tc.wantIncludes), len(e.includeFields))
			for _, field := range tc.wantIncludes {
				assert.NotNil(t, e.includeFields[field.FQDN()])
			}

			require.Equal(t, len(tc.wantExcludes), len(e.excludeFields))
			for _, field := range tc.wantExcludes {
				assert.NotNil(t, e.excludeFields[field.FQDN()])
			}

			d, err := json.Marshal(e)
			require.NoError(t, err)
			require.NotNil(t, d)

			e3 := new(Endpoint)
			err = json.Unmarshal(d, e3)
			require.NoError(t, err)
			assert.Equal(t, e.Version, e3.Version)
			assert.Equal(t, e.Type, e3.Type)
			assert.Equal(t, e.CalledAs, e3.CalledAs)
			assert.Equal(t, e.Path, e3.Path)
			assert.Equal(t, e.Description, e3.Description)
			assert.Equal(t, e.Datasource, e3.Datasource)
			assert.Equal(t, e.Table, e3.Table)
			assert.Equal(t, e.Fields, e3.Fields)
			assert.Equal(t, e.Filter, e3.Filter)

			d, err = yaml.Marshal(e)
			require.NoError(t, err)
			require.NotNil(t, d)

			e4 := new(Endpoint)
			err = yaml.Unmarshal(d, e4)
			require.NoError(t, err)
			assert.Equal(t, e.Version, e4.Version)
			assert.Equal(t, e.Type, e4.Type)
			assert.Equal(t, e.CalledAs, e4.CalledAs)
			assert.Equal(t, e.Path, e4.Path)
			assert.Equal(t, e.Description, e4.Description)
			assert.Equal(t, e.Datasource, e4.Datasource)
			assert.Equal(t, e.Table, e4.Table)
			assert.Equal(t, e.Fields, e4.Fields)
			assert.Equal(t, e.Filter, e4.Filter)
		})
	}
}

func TestEndpoint_JSON_error(t *testing.T) {
	t.Parallel()

	t.Run("field json error", func(t *testing.T) {
		t.Parallel()

		b := bytes.NewBufferString("\000\001")

		f := new(Endpoint)
		err := f.UnmarshalJSON(b.Bytes())
		require.Error(t, err)
	})
}

func TestEndpoint_YAML_error(t *testing.T) {
	t.Parallel()

	t.Run("field yaml error", func(t *testing.T) {
		t.Parallel()

		b := bytes.NewBufferString("\000\001")

		f := new(Endpoint)
		err := f.UnmarshalYAML(b.Bytes())
		require.Error(t, err)
	})
}

func TestEndpoint_FieldIncluded(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		excludes []string
		includes []string
		fqid     string
		want     bool
	}{
		{name: "no match", fqid: "abc", want: true},
		{name: "excluded", excludes: []string{"abc"}, fqid: "abc"},
		{name: "included", includes: []string{"abc"}, fqid: "abc", want: true},
		{name: "excluded & included", excludes: []string{"abc"}, includes: []string{"abc"}, fqid: "abc"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := &Endpoint{
				includeFields: make(map[string]*Field),
				excludeFields: make(map[string]*Field),
			}

			for i := range tc.includes {
				e.includeFields[tc.includes[i]] = &Field{}
			}
			for i := range tc.excludes {
				e.excludeFields[tc.excludes[i]] = &Field{}
			}

			got := e.FieldIncluded(tc.fqid)
			assert.Equal(t, tc.want, got)
		})
	}
}
