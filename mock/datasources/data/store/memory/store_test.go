package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("new", func(t *testing.T) {
		t.Parallel()

		s := New(mockFDS)
		require.NotNil(t, s)

		s2, ok := s.(*storage)
		require.True(t, ok)
		require.NotNil(t, s2)
		require.NotNil(t, s2.sourceDefs)
		require.NotNil(t, s2.sources)
		require.NotNil(t, s2.tables)

		assert.Equal(t, 2, len(s2.sourceDefs))
		assert.Equal(t, 2, len(s2.sources))
	})
}

func TestStorage_FindTableDef(t *testing.T) {
	t.Parallel()

	s := New(mockFDS)
	require.NotNil(t, s)

	s2, ok := s.(*storage)
	require.True(t, ok)
	require.NotNil(t, s2)

	testCases := []struct {
		name    string
		source  string
		table   string
		wantErr bool
	}{
		{name: "empty source", table: "personen", wantErr: true},
		{name: "unknown source", source: "huh", table: "personen", wantErr: true},
		{name: "empty table", source: "brp", table: "", wantErr: true},
		{name: "unknown table", source: "brp", table: "kenteken", wantErr: true},
		{name: "brp personen", source: "BRP", table: "Persoon"},
		{name: "rdw kentekens", source: "RDW", table: "Kenteken"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := s2.findTableDef(tc.source, tc.table)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
			}
		})
	}
}

func TestStorage_FindTable(t *testing.T) {
	t.Parallel()

	s := New(mockFDS)
	require.NotNil(t, s)

	s2, ok := s.(*storage)
	require.True(t, ok)
	require.NotNil(t, s2)

	testCases := []struct {
		name    string
		source  string
		table   string
		wantErr bool
	}{
		{name: "empty source", table: "personen", wantErr: true},
		{name: "unknown source", source: "huh", table: "personen", wantErr: true},
		{name: "empty table", source: "brp", table: "", wantErr: true},
		{name: "unknown table", source: "brp", table: "kenteken", wantErr: true},
		{name: "brp personen", source: "BRP", table: "Persoon"},
		{name: "rdw kentekens", source: "RDW", table: "Kenteken"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := s2.findTable(tc.source, tc.table)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
			}
		})
	}
}

func TestStorage_FindUnqualifiedTable(t *testing.T) {
	t.Parallel()

	s := New(mockFDS)
	require.NotNil(t, s)

	s2, ok := s.(*storage)
	require.True(t, ok)
	require.NotNil(t, s2)

	testCases := []struct {
		name    string
		table   string
		wantErr bool
	}{
		{name: "empty table", table: "", wantErr: true},
		{name: "unknown table", table: "adressen", wantErr: true},
		{name: "duplicate", table: "adres", wantErr: true},
		{name: "brp personen", table: "Persoon"},
		{name: "rdw kentekens", table: "Kenteken"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := s2.findUnqualifiedTable(tc.table)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
			}
		})
	}
}
