package models

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

var datasource1 = &schema.Datasource{
	Parent:      schema.Parent{ID: "src1"},
	Description: "source 1",
	Tables:      []*schema.Table{table1},
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

		_ = datasource1.Table("") // force internal fix().

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

		_ = datasource1.Table("") // force internal fix().

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

func TestDatasource_MatchFilter(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		filter map[string]any
		want   bool
	}{
		{name: "no filter", want: true},
		{name: "all filter", filter: map[string]any{"id": regexp.MustCompile(".*")}, want: true},
		{name: "bad field", filter: map[string]any{"table": "src1"}},
		{name: "id mismatch", filter: map[string]any{"id": "src2"}},
		{name: "id match", filter: map[string]any{"id": "src1"}, want: true},
		{name: "id regex mismatch", filter: map[string]any{"id": regexp.MustCompile("srd.*")}},
		{name: "id regex match", filter: map[string]any{"id": regexp.MustCompile("src.*")}, want: true},
		{name: "description mismatch", filter: map[string]any{"description": "source 2"}},
		{name: "description match", filter: map[string]any{"description": "source 1"}, want: true},
		{name: "description regex mismatch", filter: map[string]any{"description": regexp.MustCompile("soup.*")}},
		{name: "description regex match", filter: map[string]any{"description": regexp.MustCompile("sour.*")}, want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s1 := NewSource(datasource1)
			require.NotNil(t, s1)
			require.NotNil(t, s1.Tables)

			got := s1.MatchFilter(tc.filter)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDatasource_AsRecord(t *testing.T) {
	t.Parallel()

	t.Run("as record", func(t *testing.T) {
		s1 := NewSource(datasource1)
		require.NotNil(t, s1)
		require.NotNil(t, s1.Tables)

		got := s1.AsRecord()
		require.NotNil(t, got)

		assert.Len(t, got.Data, 4)
		assert.Equal(t, "src1", got.Data["fqdn"])
		assert.Equal(t, "src1", got.Data["id"])
		assert.Equal(t, "source 1", got.Data["description"])
	})
}
