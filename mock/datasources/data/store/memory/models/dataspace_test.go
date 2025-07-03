package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
)

func TestDataspace_AddDatasource(t *testing.T) {
	t.Parallel()

	t.Run("add datasource", func(t *testing.T) {
		t.Parallel()

		s1 := NewSpace(nil)

		s2 := &Datasource{
			def: &schema.Datasource{
				Parent:      schema.Parent{ID: "src1"},
				Description: "source 1",
			},
		}
		s1.AddDatasource(s2)

		got := s1.Sources["src1"]
		assert.Equal(t, s2, got)
	})
}

func TestDataspace_AsRecord(t *testing.T) {
	t.Parallel()

	t.Run("as record", func(t *testing.T) {
		s1 := NewSpace(&schema.Dataspace{
			Parent:      schema.Parent{ID: "fds"},
			Description: "FDS",
			DataSources: []*schema.Datasource{{Parent: schema.Parent{ID: "src1"}, Description: "source 1"}},
		})
		require.NotNil(t, s1)

		got := s1.AsRow()
		require.NotNil(t, got)

		assert.Len(t, got.Data, 4)
		assert.Equal(t, "fds", got.Data["fqdn"])
		assert.Equal(t, "fds", got.Data["id"])
		assert.Equal(t, "FDS", got.Data["description"])
	})
}
