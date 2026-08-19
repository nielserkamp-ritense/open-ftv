package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
)

// table1 is fixed once, as the schema is read-only once it has been fixed.
var table1 = fixedTable()

func fixedTable() *schema.Table {
	t := &schema.Table{
		Object: schema.Object{
			Parent:      schema.Parent{ID: "t1"},
			Description: "table 1",
			Fields: []*schema.Field{
				{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: enums.StringType, IsPII: true},
				{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: enums.IntegerType},
				{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: enums.BooleanType},
			},
		},
		PrimaryKey: &schema.Index{
			Parent:      schema.Parent{ID: "pk"},
			Description: "primary key",
			Fields:      []string{"f1"},
		},
		SecondaryIndexes: []*schema.Index{
			{
				Parent:      schema.Parent{ID: "ix2"},
				Description: "second index",
				Fields:      []string{"f3"},
			},
			{
				Parent:      schema.Parent{ID: "ix3"},
				Description: "third index",
				Fields:      []string{"f2"},
			},
		},
		ForeignKeys: []*schema.ForeignKey{
			{
				Index: schema.Index{
					Parent:      schema.Parent{ID: "fk1"},
					Description: "foreign key 1",
					Fields:      []string{"f1"},
				},
				ForeignTable: "foreign1",
			},
		},
	}
	t.Fix(nil)

	return t
}

func TestNewTable(t *testing.T) {
	t.Parallel()

	t.Run("new table", func(t *testing.T) {
		t.Parallel()

		t1 := newTable(table1, 0)
		require.NotNil(t, t1)
		require.NotNil(t, t1.PK)
		require.NotNil(t, t1.Indexes)
		assert.Equal(t, table1, t1.Definition())
	})
}

func TestTableFromData(t *testing.T) {
	t.Parallel()

	t.Run("table from data", func(t *testing.T) {
		t.Parallel()

		t1 := TableFromData(table1, []map[string]any{
			{"f1": "hello world", "f2": "123", "f3": true},
			{"f1": "hello mars", "f2": "321", "f3": false},
		})
		require.NotNil(t, t1)
		assert.Equal(t, table1, t1.Definition())

		assert.Equal(t, 2, len(t1.PK))
		assert.Equal(t, 2, len(t1.Indexes))
		require.NotNil(t, t1.Indexes["ix2"])
		require.NotNil(t, t1.Indexes["ix3"])
		assert.Equal(t, 2, len(t1.Indexes["ix2"]))
		assert.Equal(t, 2, len(t1.Indexes["ix3"]))
	})
}

func TestTableFromCSV(t *testing.T) {
	t.Parallel()

	t.Run("table from csv", func(t *testing.T) {
		t.Parallel()

		t1 := TableFromCSV(table1, [][]string{
			{"f1", "f2", "f3"},
			{"hello world", "123", "true"},
			{"hello mars", "321", "false"},
		})
		require.NotNil(t, t1)
		assert.Equal(t, table1, t1.Definition())

		assert.Equal(t, 2, len(t1.PK))
		assert.Equal(t, 2, len(t1.Indexes))
		require.NotNil(t, t1.Indexes["ix2"])
		require.NotNil(t, t1.Indexes["ix3"])
		assert.Equal(t, 2, len(t1.Indexes["ix2"]))
		assert.Equal(t, 2, len(t1.Indexes["ix3"]))
	})
}

func TestKeyFromData(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		keys  []any
		index *schema.Index
		want  string
	}{
		{
			name: "no key fields",
			index: &schema.Index{
				Parent:      schema.Parent{ID: "ix1"},
				Description: "ix1",
				Fields:      []string{},
			},
		},
		{
			name: "1 key field",
			keys: []any{"hello world"},
			index: &schema.Index{
				Parent:      schema.Parent{ID: "ix1"},
				Description: "ix1",
				Fields:      []string{"f1"},
			},
			want: "hello world",
		},
		{
			name: "few key fields",
			keys: []any{"hello world", 876, true},
			index: &schema.Index{
				Parent:      schema.Parent{ID: "ix1"},
				Description: "ix1",
				Fields:      []string{"f1", "f2", "f3"},
			},
			want: "hello world|876|true",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			table := &schema.Table{
				Object: schema.Object{
					Parent:      schema.Parent{ID: "t1"},
					Description: "table 1",
					Fields: []*schema.Field{
						{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}},
						{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}},
						{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}},
					},
				},
				SecondaryIndexes: []*schema.Index{tc.index},
			}

			table.Fix(nil)

			tc.index.Fix(table)

			got := KeyFromData(tc.keys, tc.index)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestTable_CreateRow(t *testing.T) {
	t.Parallel()

	t.Run("create row", func(t *testing.T) {
		t.Parallel()

		t1 := newTable(table1, 0)
		require.NotNil(t, t1)

		err := t1.CreateRow(&Row{Data: map[string]any{"f1": "123", "f2": "hello world", "f3": true}})
		require.NoError(t, err)
		require.Len(t, t1.Data, 1)
	})
}

func TestTable_DummyRecord(t *testing.T) {
	t.Parallel()

	t.Run("dummy record", func(t *testing.T) {
		t.Parallel()

		t1 := newTable(table1, 0)
		require.NotNil(t, t1)

		got := t1.DummyRecord()
		want := &Row{Data: map[string]any{"f1": nil, "f2": nil, "f3": nil}}
		assert.EqualValues(t, want.Data, got.Data)
	})

}
