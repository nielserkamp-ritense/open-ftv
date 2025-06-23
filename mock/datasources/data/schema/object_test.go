package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObject_IterateFields(t *testing.T) {
	t.Parallel()

	t.Run("object iterate fields", func(t *testing.T) {
		t.Parallel()

		o := &Object{
			Parent:      Parent{ID: "o1"},
			Description: "object 1",
			Fields: []*Field{
				{Object: Object{Parent: Parent{ID: "f1"}}},
				{Object: Object{Parent: Parent{ID: "f2"}}},
				{Object: Object{Parent: Parent{ID: "f3"}}},
				{Object: Object{Parent: Parent{ID: "f4"}}},
				{Object: Object{Parent: Parent{ID: "f5"}}},
			},
		}

		o.fixFields(nil, nil)

		var count int
		o.IterateFields(func(field *Field) {
			count++
			require.NotNil(t, field)
		})

		assert.Equal(t, 5, count)
	})
}

func TestObject_Fix(t *testing.T) {
	t.Parallel()

	f1 := &Field{Object: Object{Parent: Parent{ID: "f1"}}}
	f2 := &Field{Object: Object{Parent: Parent{ID: "f1"}}}
	f3 := &Field{Object: Object{Parent: Parent{ID: "f1"}}}

	t1 := &Table{Object: Object{Parent: Parent{ID: "t1"}, Fields: []*Field{f1, f2}}}
	t2 := &Table{Object: Object{Parent: Parent{ID: "t2"}, Fields: []*Field{f2, f3}}}

	ds := &Datasource{
		Parent:      Parent{ID: "ds1"},
		Description: "source1",
		Tables:      []*Table{t1, t2},
	}

	ds.Fix(nil)

	testCases := []struct {
		name       string
		o          *Object
		parent     *Parent
		table      *Table
		field      *Field
		wantParent *Parent
		wantTable  *Table
		wantField  *Field
	}{
		{
			name:       "only parent",
			o:          &Object{},
			parent:     &ds.Parent,
			wantParent: &ds.Parent,
		},
		{
			name:       "parent + table",
			o:          &Object{},
			parent:     &ds.Parent,
			table:      t1,
			wantParent: &ds.Parent,
			wantTable:  t1,
		},
		{
			name:       "parent + field",
			o:          &Object{},
			parent:     &ds.Parent,
			field:      f3,
			wantParent: &ds.Parent,
			wantField:  f3,
		},
		{
			name: "just fields",
			o:    &Object{Fields: []*Field{f1, f2, f3}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.o.Fix(tc.parent)
			tc.o.FixFields(tc.table, tc.field)

			assert.Equal(t, tc.parent, tc.o.parent)

			for _, field := range tc.o.Fields {
				f, ok := tc.o.fields[field.ID]
				assert.True(t, ok)
				assert.NotNil(t, f)
			}
		})
	}
}
