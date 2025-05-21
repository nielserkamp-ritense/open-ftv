package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

var mockFDS = &schema.Dataspace{
	Parent:      schema.Parent{ID: "fds"},
	Description: "Federatief Data Stelsel",
	DataSources: []*schema.Datasource{
		{
			Parent:      schema.Parent{ID: "brp"},
			Description: "Basis Registratie Personen",
			Tables: []*schema.Table{
				{
					Object: schema.Object{
						Parent:      schema.Parent{ID: "persoon"},
						Description: "personen",
						Fields: []*schema.Field{
							{
								Object: schema.Object{
									Parent:      schema.Parent{ID: "bsn"},
									Description: "Burgerservicenummer",
								},
								IsPII: true,
							},
						},
					},
					PrimaryKey: &schema.Index{
						Parent:      schema.Parent{ID: "pk"},
						Description: "primary key",
						Fields:      []string{"bsn"},
					},
				},
				{
					Object: schema.Object{
						Parent:      schema.Parent{ID: "adres"},
						Description: "adressen",
						Fields: []*schema.Field{
							{
								Object: schema.Object{
									Parent:      schema.Parent{ID: "adres-id"},
									Description: "Adres identifier",
								},
							},
						},
					},
					PrimaryKey: &schema.Index{
						Parent:      schema.Parent{ID: "pk"},
						Description: "primary key",
						Fields:      []string{"adres-id"},
					},
				},
			},
		},
		{
			Parent:      schema.Parent{ID: "rdw"},
			Description: "Rijksdienst voor het Wegverkeer",
			Tables: []*schema.Table{
				{
					Object: schema.Object{
						Parent:      schema.Parent{ID: "kenteken"},
						Description: "kentekens",
						Fields: []*schema.Field{
							{
								Object: schema.Object{
									Parent:      schema.Parent{ID: "kenteken"},
									Description: "kenteken",
								},
								IsPII: true,
							},
						},
					},
					PrimaryKey: &schema.Index{
						Parent:      schema.Parent{ID: "pk"},
						Description: "primary key",
						Fields:      []string{"kenteken"},
					},
				},
				{
					Object: schema.Object{
						Parent:      schema.Parent{ID: "adres"},
						Description: "adressen",
						Fields: []*schema.Field{
							{
								Object: schema.Object{
									Parent:      schema.Parent{ID: "adres-id"},
									Description: "Adres identifier",
								},
							},
						},
					},
					PrimaryKey: &schema.Index{
						Parent:      schema.Parent{ID: "pk"},
						Description: "primary key",
						Fields:      []string{"adres-id"},
					},
				},
			},
		},
	},
}

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
