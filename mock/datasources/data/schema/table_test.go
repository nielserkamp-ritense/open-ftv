package schema

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
)

func TestTable_Field(t *testing.T) {
	t.Parallel()

	t.Run("Table marshal yaml", func(t *testing.T) {
		t.Parallel()

		t1 := &Table{
			Object: Object{
				Parent:      Parent{ID: "t1"},
				Description: "table 1",
				Fields: []*Field{
					{
						Object: Object{Parent: Parent{ID: "id"}, Description: "primary key"},
						Type:   enums.IntegerType,
					},
					{
						Object: Object{Parent: Parent{ID: "f1"}, Description: "field 1"},
						Type:   enums.UnsignedIntegerType,
					},
					{
						Object: Object{Parent: Parent{ID: "f2"}, Description: "field 2"},
						Type:   enums.StringType,
					},
				},
			},
		}

		t1.Fix(nil)

		fqdn := t1.FQDN()
		assert.Equal(t, "t1", fqdn)

		fqid := t1.FQID()
		assert.Equal(t, "t1", fqid)

		got := t1.Field("f1")
		require.NotNil(t, got)

		got = t1.Field("f2")
		require.NotNil(t, got)

		got = t1.Field("f3")
		require.Nil(t, got)

		got = t1.Field("id")
		require.NotNil(t, got)
	})
}

func TestTable_MarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("Table marshal json", func(t *testing.T) {
		t.Parallel()

		t1 := &Table{
			Object: Object{
				Parent:      Parent{ID: "t1"},
				Description: "table 1",
				Fields: []*Field{
					{
						Object: Object{Parent: Parent{ID: "id"}, Description: "primary key"},
						Type:   enums.IntegerType,
					},
					{
						Object: Object{Parent: Parent{ID: "f1"}, Description: "field 1"},
						Type:   enums.UnsignedIntegerType,
					},
				},
			},
			PrimaryKey: &Index{
				Parent:      Parent{ID: "pk"},
				Description: "primary key",
				Fields:      []string{"id"},
			},
		}

		got, err := json.Marshal(t1)
		require.NoError(t, err)
		require.NotNil(t, got)

		want := `{"id":"t1","description":"table 1","fields":[{"id":"id","description":"primary key","type":"int"},{"id":"f1","description":"field 1","type":"uint"}],"primaryKey":{"id":"pk","description":"primary key","fields":["id"]}}`

		assert.Equal(t, want, string(got))

		var t2 Table
		err = json.Unmarshal(got, &t2)
		require.NoError(t, err)
		assert.Equal(t, t1.ID, t2.ID)
		assert.Equal(t, t1.Description, t2.Description)
		assert.Equal(t, len(t1.Fields), len(t2.Fields))
		assert.Equal(t, t1.PrimaryKey.ID, t2.PrimaryKey.ID)
		assert.Equal(t, t1.PrimaryKey.Description, t2.PrimaryKey.Description)
		assert.Equal(t, len(t1.SecondaryIndexes), len(t2.SecondaryIndexes))
		assert.Equal(t, len(t1.ForeignKeys), len(t2.ForeignKeys))
	})
}

func TestTable_UnmarshalJSON_Error(t *testing.T) {
	t.Parallel()

	t.Run("Table unmarshal json error", func(t *testing.T) {
		t.Parallel()

		d2 := new(Table)
		err := d2.UnmarshalJSON([]byte("\000\001"))
		require.Error(t, err)
	})
}

func TestTable_MarshalYAML(t *testing.T) {
	t.Parallel()

	t.Run("Table marshal yaml", func(t *testing.T) {
		t.Parallel()

		t1 := &Table{
			Object: Object{
				Parent:      Parent{ID: "t1"},
				Description: "table 1",
				Fields: []*Field{
					{
						Object: Object{Parent: Parent{ID: "id"}, Description: "primary key"},
						Type:   enums.IntegerType,
					},
					{
						Object: Object{Parent: Parent{ID: "f1"}, Description: "field 1"},
						Type:   enums.UnsignedIntegerType,
					},
					{
						Object: Object{Parent: Parent{ID: "f2"}, Description: "field 2"},
						Type:   enums.StringType,
					},
				},
			},
			PrimaryKey: &Index{
				Parent:      Parent{ID: "pk"},
				Description: "primary key",
				Fields:      []string{"id"},
			},
			SecondaryIndexes: []*Index{
				{
					Parent:      Parent{ID: "ix2"},
					Description: "index 2",
					Fields:      []string{"f1", "f2"},
					Orders:      []enums.OrderType{enums.OrderAscending},
				},
			},
			ForeignKeys: []*ForeignKey{
				{
					Index: Index{
						Parent:      Parent{ID: "fk1"},
						Description: "foreign key 1",
						Fields:      []string{"f1"},
					},
					ForeignTable: "t2",
				},
			},
		}

		got, err := yaml.Marshal(t1)
		require.NoError(t, err)
		require.NotNil(t, got)

		want := `id: t1
description: table 1
fields:
- id: id
  description: primary key
  type: int
- id: f1
  description: field 1
  type: uint
- id: f2
  description: field 2
  type: string
primaryKey:
  id: pk
  description: primary key
  fields:
  - id
secondaryIndexes:
- id: ix2
  description: index 2
  fields:
  - f1
  - f2
  orders:
  - ASC
foreignKeys:
- id: fk1
  description: foreign key 1
  foreignTable: t2
  fields:
  - f1
`

		assert.Equal(t, want, string(got))

		var t2 Table
		err = yaml.Unmarshal(got, &t2)
		require.NoError(t, err)
		assert.Equal(t, t1.ID, t2.ID)
		assert.Equal(t, t1.Description, t2.Description)
		assert.Equal(t, len(t1.Fields), len(t2.Fields))
		assert.Equal(t, t1.PrimaryKey.ID, t2.PrimaryKey.ID)
		assert.Equal(t, t1.PrimaryKey.Description, t2.PrimaryKey.Description)
		assert.Equal(t, len(t1.SecondaryIndexes), len(t2.SecondaryIndexes))
		assert.Equal(t, len(t1.ForeignKeys), len(t2.ForeignKeys))
	})
}

func TestTable_UnmarshalYAML_Error(t *testing.T) {
	t.Parallel()

	t.Run("Table unmarshal yaml error", func(t *testing.T) {
		t.Parallel()

		d2 := new(Table)
		err := d2.UnmarshalYAML([]byte("\000\001"))
		require.Error(t, err)
	})
}

func TestTable_Fix(t *testing.T) {
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

	ds.Tables = append(ds.Tables, t1)
	ds.Fix(nil)

	testCases := []struct {
		name     string
		t        *Table
		wantFQDN string
		wantFQID string
	}{
		{
			name: "no indexes, no foreign keys",
			t: &Table{
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
			},
			wantFQDN: "ds.t2",
			wantFQID: "ds.t2",
		},
		{
			name: "primary key",
			t: &Table{
				Object: Object{
					Parent:      Parent{ID: "t3"},
					Description: "main table",
					Fields: []*Field{
						{Object: Object{Parent: Parent{ID: "f1"}}, Type: enums.StringType},
						{Object: Object{Parent: Parent{ID: "f2"}}, Type: enums.IntegerType},
						{Object: Object{Parent: Parent{ID: "f3"}}, Type: enums.BooleanType},
						{Object: Object{Parent: Parent{ID: "f4"}}, Type: enums.DateType},
					},
				},
				PrimaryKey: &Index{
					Parent:      Parent{ID: "pk"},
					Description: "primary key",
					Fields:      []string{"f1", "f2"},
				},
			},
			wantFQDN: "ds.t3",
			wantFQID: "ds.t3",
		},
		{
			name: "secondary indexes",
			t: &Table{
				Object: Object{
					Parent:      Parent{ID: "t4"},
					Description: "main table",
					Fields: []*Field{
						{Object: Object{Parent: Parent{ID: "f1"}}, Type: enums.StringType},
						{Object: Object{Parent: Parent{ID: "f2"}}, Type: enums.IntegerType},
						{Object: Object{Parent: Parent{ID: "f3"}}, Type: enums.BooleanType},
						{Object: Object{Parent: Parent{ID: "f4"}}, Type: enums.DateType},
					},
				},
				SecondaryIndexes: []*Index{
					{Parent: Parent{ID: "ix2"}, Description: "second index", Fields: []string{"f2"}},
					{Parent: Parent{ID: "ix3"}, Description: "third index", Fields: []string{"f4"}},
				},
			},
			wantFQDN: "ds.t4",
			wantFQID: "ds.t4",
		},
		{
			name: "foreign key",
			t: &Table{
				Object: Object{
					Parent:      Parent{ID: "t5"},
					Description: "main table",
					Fields: []*Field{
						{Object: Object{Parent: Parent{ID: "f1"}}, Type: enums.StringType},
						{Object: Object{Parent: Parent{ID: "f2"}}, Type: enums.IntegerType},
						{Object: Object{Parent: Parent{ID: "f3"}}, Type: enums.BooleanType},
						{Object: Object{Parent: Parent{ID: "f4"}}, Type: enums.DateType},
					},
				},
				ForeignKeys: []*ForeignKey{
					{
						Index: Index{
							Parent:      Parent{ID: "fk1"},
							Description: "link with table 1",
							Fields:      []string{"f1", "f3"},
						},
						ForeignTable: "t1",
					},
				},
			},
			wantFQDN: "ds.t5",
			wantFQID: "ds.t5",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			table := tc.t
			table.Fix(ds)

			assert.Equal(t, &ds.Parent, table.parent)
			assert.Equal(t, ds, table.parentSource)
			assert.Equal(t, len(table.fields), len(table.Fields))
			assert.Equal(t, len(table.indexes), len(table.SecondaryIndexes))
			assert.Equal(t, len(table.foreignKeys), len(table.ForeignKeys))
			assert.Equal(t, tc.wantFQDN, table.FQDN())
			assert.Equal(t, tc.wantFQID, table.FQID())

			if table.PrimaryKey != nil {
				assert.Equal(t, table.PrimaryKey.ID, table.PrimaryKey.ID)
				assert.Equal(t, table.PrimaryKey.Description, table.PrimaryKey.Description)
				assert.Equal(t, len(table.PrimaryKey.Fields), len(table.PrimaryKey.Fields))
			}

			if tc.t.SecondaryIndexes != nil {
				assert.Equal(t, tc.t.SecondaryIndexes, table.SecondaryIndexes)

				for _, index := range tc.t.SecondaryIndexes {
					ix := table.SecondaryIndex(index.ID)
					require.NotNil(t, ix)
				}
			}

			if tc.t.ForeignKeys != nil {
				assert.Equal(t, tc.t.ForeignKeys, table.ForeignKeys)

				for _, key := range tc.t.ForeignKeys {
					fk := table.ForeignKey(key.ID)
					require.NotNil(t, fk)
				}
			}
		})
	}
}

func TestTable_FindForeignKey(t *testing.T) {
	t.Parallel()

	pk1 := &Index{Parent: Parent{ID: "pk"}, Fields: []string{"bsn"}}

	fk1 := &ForeignKey{
		Index:        Index{Parent: Parent{ID: "fk1"}, Description: "foreign key 1", Fields: []string{"postcode"}},
		ForeignTable: "t1",
	}

	fk2 := &ForeignKey{
		Index:        Index{Parent: Parent{ID: "fk2"}, Description: "foreign key 2", Fields: []string{"bsn"}},
		ForeignTable: "t2",
	}

	fk3 := &ForeignKey{
		Index:        Index{Parent: Parent{ID: "fk3"}, Description: "foreign key 3", Fields: []string{"bsn"}},
		ForeignTable: "t1",
	}

	t1 := &Table{Object: Object{Parent: Parent{ID: "t1"}}, PrimaryKey: pk1}
	t2 := &Table{Object: Object{Parent: Parent{ID: "t2"}}}
	t3 := &Table{Object: Object{Parent: Parent{ID: "t3"}}, ForeignKeys: []*ForeignKey{fk1, fk2}}
	t4 := &Table{Object: Object{Parent: Parent{ID: "t4"}}, ForeignKeys: []*ForeignKey{fk1, fk2, fk3}}

	ds := &Datasource{
		Parent:      Parent{ID: "ds1"},
		Description: "source1",
		Tables:      []*Table{t1, t2, t3, t4},
	}
	ds.Fix(nil)

	testCases := []struct {
		name    string
		t       *Table
		foreign *Table
		want    *ForeignKey
	}{
		{name: "no foreign keys", t: t2, foreign: t1},
		{name: "foreign keys not matched", t: t3, foreign: t1},
		{name: "foreign key matched", t: t4, foreign: t1, want: fk3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.t.FindForeignKey(tc.foreign)
			assert.Equal(t, tc.want, got)
		})
	}
}
