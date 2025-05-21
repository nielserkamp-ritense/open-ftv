package memory

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/models"
)

func TestStorage_AddTableFromData(t *testing.T) {
	t.Parallel()

	t.Run("add table from data", func(t *testing.T) {
		t.Parallel()

		t1 := &schema.Table{
			Object: schema.Object{
				Parent:      schema.Parent{ID: "tab1"},
				Description: "table 1",
			},
		}

		s1 := &schema.Datasource{
			Parent: schema.Parent{ID: "SRC1"},
			Tables: []*schema.Table{t1},
		}

		s2 := &storage{
			sourceDefs: make(map[string]*schema.Datasource),
			sources:    make(map[string]*models.Datasource),
			tables:     make(map[string]*models.Table),
		}

		s2.AddDatasource(s1)

		err := s2.AddTableFromData(s1.ID, t1.ID, []map[string]any{})
		require.NoError(t, err)

		err = s2.AddTableFromData(s1.ID, "oops", []map[string]any{})
		require.Error(t, err)
	})
}

func TestStorage_AddTableFromCSV(t *testing.T) {
	t.Parallel()

	t.Run("add table from data", func(t *testing.T) {
		t.Parallel()

		t1 := &schema.Table{
			Object: schema.Object{
				Parent:      schema.Parent{ID: "tab1"},
				Description: "table 1",
			},
		}

		s1 := &schema.Datasource{
			Parent: schema.Parent{ID: "SRC1"},
			Tables: []*schema.Table{t1},
		}

		s2 := &storage{
			sourceDefs: make(map[string]*schema.Datasource),
			sources:    make(map[string]*models.Datasource),
			tables:     make(map[string]*models.Table),
		}

		s2.AddDatasource(s1)

		err := s2.AddTableFromCSV(s1.ID, t1.ID, [][]string{})
		require.NoError(t, err)

		err = s2.AddTableFromCSV(s1.ID, "oops", [][]string{})
		require.Error(t, err)
	})
}
