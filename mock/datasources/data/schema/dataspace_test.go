package schema

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataSpace_Source(t *testing.T) {
	t.Parallel()

	t.Run("dataspace source", func(t *testing.T) {
		t.Parallel()

		d := &Dataspace{
			Parent:      Parent{ID: "myspace"},
			Description: "multi-space",
			DataSources: []*Datasource{
				{Parent: Parent{ID: "ds1"}},
				{Parent: Parent{ID: "ds2"}},
				{Parent: Parent{ID: "ds3"}},
			},
		}
		d.Fix()

		got := d.Source("ds4")
		require.Nil(t, got)

		got = d.Source("ds3")
		require.NotNil(t, got)

		got = d.Source("ds2")
		require.NotNil(t, got)

		got = d.Source("ds1")
		require.NotNil(t, got)
	})
}

func TestDataSpace_MarshalJSON(t *testing.T) {
	t.Parallel()

	d1 := &Dataspace{
		Parent:      Parent{ID: "myspace"},
		Description: "multi-space",
		DataSources: []*Datasource{
			{Parent: Parent{ID: "ds1"}},
			{Parent: Parent{ID: "ds2"}},
			{Parent: Parent{ID: "ds3"}},
		},
	}

	d2 := &Dataspace{
		Parent:      Parent{ID: "empty"},
		Description: "empty-space",
	}

	want1 := `{"id":"myspace","description":"multi-space","dataSources":[{"id":"ds1"},{"id":"ds2"},{"id":"ds3"}]}`
	want2 := `{"id":"empty","description":"empty-space"}`

	testCases := []struct {
		name string
		d    *Dataspace
		want string
	}{
		{name: "with sources", d: d1, want: want1},
		{name: "without sources", d: d2, want: want2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(tc.d)
			require.NoError(t, err)
			require.NotNil(t, got)

			assert.Equal(t, tc.want, string(got))

			var got2 Dataspace
			err = json.Unmarshal(got, &got2)
			require.NoError(t, err)
			assert.Equal(t, tc.d.ID, got2.ID)
			assert.Equal(t, tc.d.Description, got2.Description)
			assert.Equal(t, len(tc.d.DataSources), len(got2.DataSources))
		})
	}
}

func TestDataSpace_MarshalYAML(t *testing.T) {
	t.Parallel()

	d1 := &Dataspace{
		Parent:      Parent{ID: "myspace"},
		Description: "multi-space",
		DataSources: []*Datasource{
			{Parent: Parent{ID: "ds1"}},
			{Parent: Parent{ID: "ds2"}},
			{Parent: Parent{ID: "ds3"}},
		},
	}

	d2 := &Dataspace{
		Parent:      Parent{ID: "empty"},
		Description: "empty-space",
	}

	want1 := `id: myspace
description: multi-space
dataSources:
- id: ds1
- id: ds2
- id: ds3
`

	want2 := `id: empty
description: empty-space
`

	testCases := []struct {
		name string
		d    *Dataspace
		want string
	}{
		{name: "with sources", d: d1, want: want1},
		{name: "without sources", d: d2, want: want2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := yaml.Marshal(tc.d)
			require.NoError(t, err)
			require.NotNil(t, got)

			assert.Equal(t, tc.want, string(got))

			var got2 Dataspace
			err = yaml.Unmarshal(got, &got2)
			require.NoError(t, err)
			assert.Equal(t, tc.d.ID, got2.ID)
			assert.Equal(t, tc.d.Description, got2.Description)
			assert.Equal(t, len(tc.d.DataSources), len(got2.DataSources))
		})
	}
}

func TestDataSpace_UnmarshalJSON_Error(t *testing.T) {
	t.Parallel()

	t.Run("dataspace unmarshal json error", func(t *testing.T) {
		t.Parallel()

		d2 := new(Dataspace)
		err := d2.UnmarshalJSON([]byte("\000\001"))
		require.Error(t, err)
	})
}

func TestDataSpace_UnmarshalYAML_Error(t *testing.T) {
	t.Parallel()

	t.Run("dataspace unmarshal yaml error", func(t *testing.T) {
		t.Parallel()

		d2 := new(Dataspace)
		err := d2.UnmarshalYAML([]byte("\000\001"))
		require.Error(t, err)
	})
}

func TestDataSpace_Fix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		d              *Dataspace
		wantFQDN       string
		wantSourceFQDN []string
	}{
		{
			name:     "no sources",
			d:        &Dataspace{Parent: Parent{ID: "empty"}, Description: "Empty space"},
			wantFQDN: "empty",
		},
		{
			name: "all",
			d: &Dataspace{
				Parent:      Parent{ID: "mine"},
				Description: "My space",
				DataSources: []*Datasource{
					{Parent: Parent{ID: "ds1"}},
					{Parent: Parent{ID: "ds2"}},
					{Parent: Parent{ID: "ds3"}},
				},
			},
			wantFQDN:       "mine",
			wantSourceFQDN: []string{"mine.ds1", "mine.ds2", "mine.ds3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := tc.d
			d.Fix()

			assert.Nil(t, d.parent)

			assert.Equal(t, len(d.DataSources), len(d.DataSources))
			assert.Equal(t, tc.wantFQDN, d.FQDN())

			for id, ds := range d.sources {
				assert.Equal(t, id, ds.ID)
			}

			for i, ds := range d.DataSources {
				assert.Equal(t, tc.wantSourceFQDN[i], ds.FQDN())
			}
		})
	}
}
