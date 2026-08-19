package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
)

var datasource1 = fixedDatasource()

func fixedDatasource() *schema.Datasource {
	d := &schema.Datasource{
		Parent:      schema.Parent{ID: "src1"},
		Description: "source 1",
		Tables:      []*schema.Table{table1},
	}
	d.Fix(nil)

	return d
}

func TestNewSource(t *testing.T) {
	t.Parallel()

	t.Run("add datasource", func(t *testing.T) {
		t.Parallel()

		s1 := NewSource(datasource1)
		require.NotNil(t, s1)
		require.NotNil(t, s1.Tables)
	})
}

func TestDatasource_AddTableFromData(t *testing.T) {
	t.Parallel()

	t.Run("new table", func(t *testing.T) {
		t.Parallel()

		s1 := NewSource(datasource1)
		require.NotNil(t, s1)
		require.NotNil(t, s1.Tables)

		s1.AddTableFromData(table1, []map[string]any{
			{"f1": "hello world", "f2": "123", "f3": true},
			{"f1": "hello mars", "f2": "321", "f3": false},
		})

		t1 := s1.Tables["t1"]
		require.NotNil(t, t1)

		assert.Equal(t, 2, len(t1.PK))
		assert.Equal(t, 2, len(t1.Indexes))
		require.NotNil(t, t1.Indexes["ix2"])
		require.NotNil(t, t1.Indexes["ix3"])
		assert.Equal(t, 2, len(t1.Indexes["ix2"]))
		assert.Equal(t, 2, len(t1.Indexes["ix3"]))
	})
}

func TestDatasource_AddTableFromCSV(t *testing.T) {
	t.Parallel()

	t.Run("new table", func(t *testing.T) {
		t.Parallel()

		s1 := NewSource(datasource1)
		require.NotNil(t, s1)
		require.NotNil(t, s1.Tables)

		s1.AddTableFromCSV(table1, [][]string{
			{"f1", "f2", "f3"},
			{"hello world", "123", "true"},
			{"hello mars", "321", "false"},
		})

		t1 := s1.Tables["t1"]
		require.NotNil(t, t1)

		assert.Equal(t, 2, len(t1.PK))
		assert.Equal(t, 2, len(t1.Indexes))
		require.NotNil(t, t1.Indexes["ix2"])
		require.NotNil(t, t1.Indexes["ix3"])
		assert.Equal(t, 2, len(t1.Indexes["ix2"]))
		assert.Equal(t, 2, len(t1.Indexes["ix3"]))
	})
}

func TestDatasource_AsRecord(t *testing.T) {
	t.Parallel()

	t.Run("as record", func(t *testing.T) {
		s1 := NewSource(datasource1)
		require.NotNil(t, s1)
		require.NotNil(t, s1.Tables)

		got := s1.AsRow()
		require.NotNil(t, got)

		assert.Len(t, got.Data, 4)
		assert.Equal(t, "src1", got.Data["fqdn"])
		assert.Equal(t, "src1", got.Data["id"])
		assert.Equal(t, "source 1", got.Data["description"])
	})
}
