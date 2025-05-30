package schema

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/types"
)

func TestForeignKey_MarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("ForeignKey marshal json", func(t *testing.T) {
		t.Parallel()

		d := &ForeignKey{
			Index: Index{
				Parent:      Parent{ID: "table2"},
				Description: "Some table",
				Fields:      []string{"f1", "f2", "f3"},
			},
			ForeignTable: "table1",
		}

		got, err := json.Marshal(d)
		require.NoError(t, err)
		require.NotNil(t, got)

		want := `{"id":"table2","description":"Some table","foreignTable":"table1","fields":["f1","f2","f3"]}`

		assert.Equal(t, want, string(got))

		var d2 ForeignKey
		err = json.Unmarshal(got, &d2)
		require.NoError(t, err)
		assert.Equal(t, d.ID, d2.ID)
		assert.Equal(t, d.Description, d2.Description)
		assert.EqualValues(t, d.Fields, d2.Fields)
		assert.Equal(t, d.ForeignTable, d2.ForeignTable)
	})
}

func TestForeignKey_UnmarshalJSON_Error(t *testing.T) {
	t.Parallel()

	t.Run("ForeignKey unmarshal json error", func(t *testing.T) {
		t.Parallel()

		d2 := new(ForeignKey)
		err := d2.UnmarshalJSON([]byte("\000\001"))
		require.Error(t, err)
	})
}

func TestForeignKey_MarshalYAML(t *testing.T) {
	t.Parallel()

	t.Run("ForeignKey marshal yaml", func(t *testing.T) {
		t.Parallel()

		d := &ForeignKey{
			Index: Index{
				Parent:      Parent{ID: "table2"},
				Description: "Some table",
				Fields:      []string{"f1", "f2", "f3"},
			},
			ForeignTable: "table1",
		}

		got, err := yaml.Marshal(d)
		require.NoError(t, err)
		require.NotNil(t, got)

		want := `id: table2
description: Some table
foreignTable: table1
fields:
- f1
- f2
- f3
`

		assert.Equal(t, want, string(got))

		var d2 ForeignKey
		err = yaml.Unmarshal(got, &d2)
		require.NoError(t, err)
		assert.Equal(t, d.ID, d2.ID)
		assert.Equal(t, d.Description, d2.Description)
		assert.EqualValues(t, d.Fields, d2.Fields)
		assert.Equal(t, d.ForeignTable, d2.ForeignTable)
	})
}

func TestForeignKey_UnmarshalYAML_Error(t *testing.T) {
	t.Parallel()

	t.Run("ForeignKey unmarshal yaml error", func(t *testing.T) {
		t.Parallel()

		d2 := new(ForeignKey)
		err := d2.UnmarshalYAML([]byte("\000\001"))
		require.Error(t, err)
	})
}

func TestForeignKey_Fix(t *testing.T) {
	t.Parallel()

	ds := &Datasource{
		Parent:      Parent{ID: "ds"},
		Description: "datasource1",
		Tables:      make([]*Table, 0),
	}

	t1 := &Table{
		Object: Object{
			Parent:      Parent{ID: "t1"},
			Description: "foreign table",
			Fields: []*Field{
				{Object: Object{Parent: Parent{ID: "f1"}}, Type: types.StringType},
				{Object: Object{Parent: Parent{ID: "f2"}}, Type: types.IntegerType},
				{Object: Object{Parent: Parent{ID: "f3"}}, Type: types.BooleanType},
				{Object: Object{Parent: Parent{ID: "f4"}}, Type: types.DateType},
			},
		},
	}

	t2 := &Table{
		Object: Object{
			Parent:      Parent{ID: "t2"},
			Description: "main table",
			Fields: []*Field{
				{Object: Object{Parent: Parent{ID: "f1"}}, Type: types.StringType},
				{Object: Object{Parent: Parent{ID: "f2"}}, Type: types.IntegerType},
				{Object: Object{Parent: Parent{ID: "f3"}}, Type: types.BooleanType},
				{Object: Object{Parent: Parent{ID: "f4"}}, Type: types.DateType},
			},
		},
	}

	ds.Tables = append(ds.Tables, t1, t2)
	ds.fix(nil)

	testCases := []struct {
		name     string
		fk       *ForeignKey
		wantFQDN string
	}{
		{
			name: "no sources",
			fk: &ForeignKey{
				Index: Index{
					Parent:      Parent{ID: "fk1"},
					Description: "Foreign key 1",
					Fields:      []string{"f2", "f1"},
				},
				ForeignTable: "t1",
			},
			wantFQDN: "ds.t2.fk1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fk := tc.fk
			fk.fix(t2, ds.tables)

			assert.Equal(t, &t2.Parent, fk.parent)
			assert.Equal(t, t2, fk.parentTable)
			assert.Equal(t, t1, fk.foreignTable)
			assert.Equal(t, len(fk.fields), len(fk.Fields))
			assert.Equal(t, tc.wantFQDN, fk.FQDN())

			for i, f := range fk.fields {
				require.NotNil(t, f)
				assert.Equal(t, fk.Fields[i], f.ID)
			}
		})
	}
}
