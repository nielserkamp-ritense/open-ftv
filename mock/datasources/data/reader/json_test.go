package reader

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
)

func TestDataspaceFromJSON(t *testing.T) {
	f1 := bytes.NewBufferString(`// oopsie \0`)

	f2 := bytes.NewBufferString(`{"id":"1","description":"first encounter"}`)

	f3 := bytes.NewBufferString(`
{
 "id": "3",
 "description": "multi-space",
 "dataSources": [{"id": "ds1"}, {"id": "ds2"}, {"id": "ds3"}]
}`)

	testCases := []struct {
		name    string
		f       io.Reader
		wantErr bool
		want    *schema.Dataspace
	}{
		{
			name:    "bad file",
			f:       f1,
			wantErr: true,
		},
		{
			name: "no sources",
			f:    f2,
			want: &schema.Dataspace{
				Parent:      schema.Parent{ID: "1"},
				Description: "first encounter",
			},
		},
		{
			name: "with sources",
			f:    f3,
			want: &schema.Dataspace{
				Parent:      schema.Parent{ID: "3"},
				Description: "multi-space",
				DataSources: []*schema.Datasource{
					{Parent: schema.Parent{ID: "ds1"}},
					{Parent: schema.Parent{ID: "ds2"}},
					{Parent: schema.Parent{ID: "ds3"}},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d, err := DataspaceFromJSON(tc.f)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, d)
			} else {
				require.NoError(t, err)
				require.NotNil(t, d)

				assert.Equal(t, tc.want.ID, d.ID)
				assert.Equal(t, tc.want.Description, d.Description)
				assert.Equal(t, len(tc.want.DataSources), len(d.DataSources))
			}
		})
	}
}

func TestDatasourceFromJSON(t *testing.T) {
	f1 := bytes.NewBufferString(`// oopsie \0`)

	f2 := bytes.NewBufferString(`{"id":"1","description":"first encounter"}`)

	f3 := bytes.NewBufferString(`
{
 "id": "3",
 "description": "multi-space",
 "tables": [{"id": "t1"}, {"id": "t2"}, {"id": "t3"}]
}`)

	testCases := []struct {
		name    string
		f       io.Reader
		wantErr bool
		want    *schema.Datasource
	}{
		{
			name:    "bad file",
			f:       f1,
			wantErr: true,
		},
		{
			name: "no tables",
			f:    f2,
			want: &schema.Datasource{
				Parent:      schema.Parent{ID: "1"},
				Description: "first encounter",
			},
		},
		{
			name: "with tables",
			f:    f3,
			want: &schema.Datasource{
				Parent:      schema.Parent{ID: "3"},
				Description: "multi-space",
				Tables: []*schema.Table{
					{Object: schema.Object{Parent: schema.Parent{ID: "t1"}}},
					{Object: schema.Object{Parent: schema.Parent{ID: "t2"}}},
					{Object: schema.Object{Parent: schema.Parent{ID: "t3"}}},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d, err := DatasourceFromJSON(tc.f)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, d)
			} else {
				require.NoError(t, err)
				require.NotNil(t, d)

				assert.Equal(t, tc.want.ID, d.ID)
				assert.Equal(t, tc.want.Description, d.Description)
				assert.Equal(t, len(tc.want.Tables), len(d.Tables))
			}
		})
	}
}
