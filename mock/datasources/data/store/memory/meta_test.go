package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/reader"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/models"
)

func TestStorage_SetDataspace(t *testing.T) {
	t.Parallel()

	t.Run("set dataspace", func(t *testing.T) {
		t.Parallel()

		s1 := &schema.Dataspace{
			Parent:      schema.Parent{ID: "fds"},
			Description: "FDS",
			DataSources: []*schema.Datasource{
				{Parent: schema.Parent{ID: "brp"}, Description: "BRP"},
				{Parent: schema.Parent{ID: "rdw"}, Description: "RDW"},
			},
		}

		s2 := New(nil)
		require.NotNil(t, s2)

		s3, ok := s2.(*storage)
		require.True(t, ok)
		require.NotNil(t, s3)

		s2.SetDataspace(s1)
		assert.Equal(t, s1, s3.spaceDef)

		s4 := s2.GetDataspace()
		require.NotNil(t, s4)
	})
}

func TestStorage_AddDatasource(t *testing.T) {
	t.Parallel()

	t.Run("add datasource", func(t *testing.T) {
		t.Parallel()

		s1 := &schema.Datasource{Parent: schema.Parent{ID: "SRC1"}}
		s2 := &storage{
			sourceDefs: make(map[string]*schema.Datasource),
			sources:    make(map[string]*models.Datasource),
		}

		s2.AddDatasource(s1)

		got := s2.sourceDefs["src1"]
		assert.Equal(t, s1, got)

		got2 := s2.sources["src1"]
		assert.NotNil(t, got2)

		got3 := s2.GetDatasources()
		assert.NotNil(t, got3)
		assert.Len(t, got3, 1)
	})
}

func TestStorage_AddEndpoint(t *testing.T) {
	t.Parallel()

	t.Run("add endpoint", func(t *testing.T) {
		t.Parallel()

		e1 := &schema.Endpoint{
			Version:     1,
			Type:        1,
			Path:        "/path/",
			FullVersion: "1.0.0",
			Description: "my endpoint",
			Datasource:  "ds1",
			Table:       "t1",
		}

		s1 := &storage{endpoints: make(map[string]*schema.Endpoint)}
		s1.AddEndpoint(e1)

		got := s1.endpoints["/v1/path"]
		assert.Equal(t, e1, got)

		var count int
		s1.IterateEndpoints(func(def *schema.Endpoint) {
			count++
		})
		assert.Equal(t, 1, count)
	})
}

func TestStorage_DatasourceExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		id   string
		want bool
	}{
		{name: "empty"},
		{name: "unknown", id: "RDW"},
		{name: "found", id: "brp", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New(nil)
			require.NotNil(t, s)

			err := reader.LoadFromPath(s, "../../../../../testdata/dataspaces/fds")
			require.NoError(t, err)

			got := s.DatasourceExists(tc.id)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestStorage_TableExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		id   string
		want bool
	}{
		{name: "empty"},
		{name: "unknown", id: "brp.hello"},
		{name: "found", id: "brp.adres", want: true},
		{name: "many qualifiers", id: "x.y.z.adres"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New(nil)
			require.NotNil(t, s)

			err := reader.LoadFromPath(s, "../../../../../testdata/dataspaces/fds")
			require.NoError(t, err)

			got := s.TableExists(tc.id)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestStorage_GetDatasource(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		id   string
		want bool
	}{
		{name: "empty"},
		{name: "unknown", id: "brk"},
		{name: "found", id: "BRP", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New(nil)
			require.NotNil(t, s)

			err := reader.LoadFromPath(s, "../../../../../testdata/dataspaces/fds")
			require.NoError(t, err)

			got := s.GetDatasource(tc.id)
			assert.Equal(t, tc.want, got != nil)
		})
	}
}

func TestStorage_GetTable(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		id   string
		want bool
	}{
		{name: "empty"},
		{name: "unknown", id: "brp.hello"},
		{name: "found unqualified", id: "adres", want: true},
		{name: "found qualified", id: "brp.adres", want: true},
		{name: "found fully qualified", id: "fds.brp.adres", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New(nil)
			require.NotNil(t, s)

			err := reader.LoadFromPath(s, "../../../../../testdata/dataspaces/fds")
			require.NoError(t, err)

			got, err2 := s.GetTable(tc.id)
			if tc.want {
				require.Nil(t, err2)
				require.NotNil(t, got)
			} else {
				require.NotNil(t, err2)
				require.Nil(t, got)
			}
		})
	}
}
