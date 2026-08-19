package schema

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataSource_Table(t *testing.T) {
	t.Parallel()

	t.Run("datasource table", func(t *testing.T) {
		t.Parallel()

		d := &Datasource{
			Parent:      Parent{ID: "3"},
			Description: "multi-space",
			Tables: []*Table{
				{Object: Object{Parent: Parent{ID: "t1"}}},
				{Object: Object{Parent: Parent{ID: "t2"}}},
				{Object: Object{Parent: Parent{ID: "t3"}}},
			},
		}
		d.Fix(nil)

		got := d.Table("t3")
		require.NotNil(t, got)

		got = d.Table("t2")
		require.NotNil(t, got)

		got = d.Table("t1")
		require.NotNil(t, got)

		got = d.Table("t4")
		require.Nil(t, got)
	})
}

func TestDataSource_MarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("datasource marshal json", func(t *testing.T) {
		t.Parallel()

		d := &Datasource{
			Parent:      Parent{ID: "3"},
			Description: "multi-space",
			Tables: []*Table{
				{Object: Object{Parent: Parent{ID: "t1"}}},
				{Object: Object{Parent: Parent{ID: "t2"}}},
				{Object: Object{Parent: Parent{ID: "t3"}}},
			},
		}

		got, err := json.Marshal(d)
		require.NoError(t, err)
		require.NotNil(t, got)

		want := `{"id":"3","description":"multi-space","tables":[{"id":"t1"},{"id":"t2"},{"id":"t3"}]}`

		assert.Equal(t, want, string(got))

		d2 := new(Datasource)
		err = json.Unmarshal(got, d2)
		require.NoError(t, err)

		assert.Equal(t, d.ID, d2.ID)
		assert.Equal(t, d.Description, d2.Description)
		assert.Equal(t, len(d.Tables), len(d2.Tables))
	})
}

func TestDataSource_UnmarshalJSON_Error(t *testing.T) {
	t.Parallel()

	t.Run("datasource unmarshal json error", func(t *testing.T) {
		t.Parallel()

		d2 := new(Datasource)
		err := d2.UnmarshalJSON([]byte("\000\001"))
		require.Error(t, err)
	})
}

func TestDataSource_MarshalYAML(t *testing.T) {
	t.Parallel()

	t.Run("datasource marshal yaml", func(t *testing.T) {
		t.Parallel()

		d := &Datasource{
			Parent:      Parent{ID: "3"},
			Description: "multi-space",
			Tables: []*Table{
				{Object: Object{Parent: Parent{ID: "t1"}}},
				{Object: Object{Parent: Parent{ID: "t2"}}},
				{Object: Object{Parent: Parent{ID: "t3"}}},
			},
		}

		got, err := yaml.Marshal(d)
		require.NoError(t, err)
		require.NotNil(t, got)

		want := `id: "3"
description: multi-space
tables:
- id: t1
- id: t2
- id: t3
`

		assert.Equal(t, want, string(got))

		d2 := new(Datasource)
		err = yaml.Unmarshal(got, d2)
		require.NoError(t, err)

		assert.Equal(t, d.ID, d2.ID)
		assert.Equal(t, d.Description, d2.Description)
		assert.Equal(t, len(d.Tables), len(d2.Tables))
	})
}

func TestDataSource_UnmarshalYAML_Error(t *testing.T) {
	t.Parallel()

	t.Run("datasource unmarshal yaml error", func(t *testing.T) {
		t.Parallel()

		d2 := new(Datasource)
		err := d2.UnmarshalYAML([]byte("\000\001"))
		require.Error(t, err)
	})
}

func TestDataSource_Fix(t *testing.T) {
	t.Parallel()

	ds := &Dataspace{Parent: Parent{ID: "mine"}, Description: "my space"}

	testCases := []struct {
		name          string
		d             *Datasource
		parent        *Dataspace
		wantParent    *Parent
		wantFQDN      string
		wantTableFQDN []string
	}{
		{
			name:     "no dataspace",
			d:        &Datasource{Parent: Parent{ID: "lonely"}, Description: "Lonely source"},
			wantFQDN: "lonely",
		},
		{
			name:       "no tables",
			d:          &Datasource{Parent: Parent{ID: "empty"}, Description: "Empty source"},
			parent:     ds,
			wantParent: &ds.Parent,
			wantFQDN:   "mine.empty",
		},
		{
			name: "all",
			d: &Datasource{
				Parent:      Parent{ID: "empty"},
				Description: "Empty source",
				Tables: []*Table{
					{Object: Object{Parent: Parent{ID: "t1"}}},
					{Object: Object{Parent: Parent{ID: "t2"}}},
					{Object: Object{Parent: Parent{ID: "t3"}}},
				},
			},
			parent:        ds,
			wantParent:    &ds.Parent,
			wantFQDN:      "mine.empty",
			wantTableFQDN: []string{"mine.empty.t1", "mine.empty.t2", "mine.empty.t3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := tc.d
			d.Fix(tc.parent)

			if tc.parent != nil {
				assert.Equal(t, tc.wantParent, d.parent)
				assert.Equal(t, tc.parent, d.ds)
			} else {
				assert.Nil(t, d.parent)
				assert.Nil(t, d.ds)
			}

			assert.Equal(t, len(d.Tables), len(d.tables))
			assert.Equal(t, tc.wantFQDN, d.FQDN())

			for id, table := range d.tables {
				assert.Equal(t, id, table.ID)
			}

			for i, table := range d.Tables {
				assert.Equal(t, tc.wantTableFQDN[i], table.FQDN())
			}
		})
	}
}
