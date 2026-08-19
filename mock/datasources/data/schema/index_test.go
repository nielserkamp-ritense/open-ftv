package schema

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
)

func TestIndex_IterateFields(t *testing.T) {
	t.Parallel()

	t.Run("Index iterate fields", func(t *testing.T) {
		t.Parallel()

		ix2 := &Index{
			Parent:      Parent{ID: "index2"},
			Description: "second index",
			Fields:      []string{"f1", "f2"},
			Orders:      []enums.OrderType{enums.OrderAscending, enums.OrderDescending},
		}

		t1 := &Table{
			Object: Object{
				Parent:      Parent{ID: "t1"},
				Description: "table 1",
				Fields: []*Field{
					{Object: Object{Parent: Parent{ID: "f1"}}},
					{Object: Object{Parent: Parent{ID: "f2"}}},
				},
			},
			SecondaryIndexes: []*Index{ix2},
		}

		t1.Fix(nil)

		var count int
		ix2.IterateFields(func(field *Field) {
			count++
			require.NotNil(t, field)
		})

		assert.Equal(t, 2, count)
	})
}

func TestIndex_MarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("Index marshal json", func(t *testing.T) {
		t.Parallel()

		d := &Index{
			Parent:      Parent{ID: "index2"},
			Description: "second index",
			Fields:      []string{"f1", "f2"},
			Orders:      []enums.OrderType{enums.OrderAscending, enums.OrderDescending},
		}

		got, err := json.Marshal(d)
		require.NoError(t, err)
		require.NotNil(t, got)

		want := `{"id":"index2","description":"second index","fields":["f1","f2"],"orders":["ASC","DESC"]}`

		assert.Equal(t, want, string(got))

		var d2 Index
		err = json.Unmarshal(got, &d2)
		require.NoError(t, err)
		assert.Equal(t, d.ID, d2.ID)
		assert.Equal(t, d.Description, d2.Description)
		assert.EqualValues(t, d.Fields, d2.Fields)
		assert.EqualValues(t, d.Orders, d2.Orders)
	})
}

func TestIndex_UnmarshalJSON_Error(t *testing.T) {
	t.Parallel()

	t.Run("Index unmarshal json error", func(t *testing.T) {
		t.Parallel()

		d2 := new(Index)
		err := d2.UnmarshalJSON([]byte("\000\001"))
		require.Error(t, err)
	})
}

func TestIndex_MarshalYAML(t *testing.T) {
	t.Parallel()

	t.Run("Index marshal yaml", func(t *testing.T) {
		t.Parallel()

		d := &Index{
			Parent:      Parent{ID: "index2"},
			Description: "second index",
			Fields:      []string{"f1", "f2"},
			Orders:      []enums.OrderType{enums.OrderAscending, enums.OrderDescending},
		}

		got, err := yaml.Marshal(d)
		require.NoError(t, err)
		require.NotNil(t, got)

		want := `id: index2
description: second index
fields:
- f1
- f2
orders:
- ASC
- DESC
`

		assert.Equal(t, want, string(got))

		var d2 Index
		err = yaml.Unmarshal(got, &d2)
		require.NoError(t, err)
		assert.Equal(t, d.ID, d2.ID)
		assert.Equal(t, d.Description, d2.Description)
		assert.EqualValues(t, d.Fields, d2.Fields)
		assert.EqualValues(t, d.Orders, d2.Orders)
	})
}

func TestIndex_UnmarshalYAML_Error(t *testing.T) {
	t.Parallel()

	t.Run("Index unmarshal yaml error", func(t *testing.T) {
		t.Parallel()

		d2 := new(Index)
		err := d2.UnmarshalYAML([]byte("\000\001"))
		require.Error(t, err)
	})
}

func TestIndex_Fix(t *testing.T) {
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
				{Object: Object{Parent: Parent{ID: "f1"}}, Type: enums.StringType},
				{Object: Object{Parent: Parent{ID: "f2"}}, Type: enums.IntegerType},
				{Object: Object{Parent: Parent{ID: "f3"}}, Type: enums.BooleanType},
				{Object: Object{Parent: Parent{ID: "f4"}}, Type: enums.DateType},
			},
		},
	}

	t2 := &Table{
		Object: Object{
			Parent:      Parent{ID: "t2"},
			Description: "main table",
			Fields: []*Field{
				{Object: Object{Parent: Parent{ID: "f1"}}, Type: enums.StringType},
				{Object: Object{Parent: Parent{ID: "f2"}}, Type: enums.IntegerType},
				{Object: Object{Parent: Parent{ID: "f3"}}, Type: enums.BooleanType},
				{Object: Object{Parent: Parent{ID: "f4"}}, Type: enums.DateType},
			},
		},
	}

	ds.Tables = append(ds.Tables, t1, t2)
	ds.Fix(nil)

	testCases := []struct {
		name     string
		fk       *Index
		wantFQDN string
	}{
		{
			name: "no sources",
			fk: &Index{
				Parent:      Parent{ID: "fk1"},
				Description: "Foreign key 1",
				Fields:      []string{"f2", "f1"},
				Orders:      []enums.OrderType{enums.OrderAscending, enums.OrderDescending},
			},
			wantFQDN: "ds.t2.fk1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fk := tc.fk
			fk.Fix(t2)

			assert.Equal(t, &t2.Parent, fk.parent)
			assert.Equal(t, t2, fk.parentTable)
			assert.Equal(t, len(fk.fields), len(fk.Fields))
			assert.Equal(t, tc.wantFQDN, fk.FQDN())

			for i, f := range fk.fields {
				require.NotNil(t, f)
				assert.Equal(t, fk.Fields[i], f.ID)
			}
		})
	}
}

func TestIndex_Equal(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		ix    *Index
		other []string
		want  bool
	}{
		{
			name:  "single, not equal",
			ix:    &Index{Fields: []string{"bsn"}},
			other: []string{"naam"},
		},
		{
			name:  "single, equal",
			ix:    &Index{Fields: []string{"bsn"}},
			other: []string{"bsn"},
			want:  true,
		},
		{
			name:  "multiple, not equal",
			ix:    &Index{Fields: []string{"bsn", "volgnummer"}},
			other: []string{"volgnummer", "bsn", "ingang"},
		},
		{
			name:  "multiple, equal",
			ix:    &Index{Fields: []string{"bsn", "volgnummer"}},
			other: []string{"volgnummer", "bsn"},
			want:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.ix.Equal(tc.other)
			assert.Equal(t, tc.want, got)
		})
	}
}
