package joins

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/types"
)

func TestStorage_ProcessJoins(t *testing.T) {
	t.Parallel()

	ds.Fix(nil)

	p1 := map[string]any{"bsn": "999990391", "voornaam": "Jan", "achternaam": "Jansen"}
	p2 := map[string]any{"bsn": "999990408", "voornaam": "Piet", "achternaam": "Pietersen"}
	p3 := map[string]any{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen"}

	t1 := models.TableFromData(persoon, []map[string]any{p1, p2, p3})

	t2 := models.TableFromData(adres, []map[string]any{
		{"bsn": "999990391", "postcode": "1111ZZ"},
		{"bsn": "999990408", "postcode": "2222YY"},
		{"bsn": "999990433", "postcode": "4444WW"},
	})

	t3 := models.TableFromData(kenteken, []map[string]any{})

	s := &myMeta{tables: map[string]*models.Table{"persoon": t1, "adres": t2, "kenteken": t3}}

	in1 := models.Rows{{Data: p1}, {Data: p2}, {Data: p3}}
	in2 := models.Rows{{Data: p3}}
	in3 := models.Rows{{Data: p3}, {Data: p2}}

	testCases := []struct {
		name    string
		in      models.Rows
		primary *models.Table
		join    *schema.Join
		fk      *schema.ForeignKey
		filter  map[string]any
		wantErr bool
		want    []map[string]any
	}{
		{
			name:    "bad foreign key",
			in:      in1,
			primary: t3,
			join:    &schema.Join{Type: types.ForcedParentChild, Source: "adres"},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk3"}, Fields: []string{"postcode"}},
				ForeignTable: "kenteken",
			},
			wantErr: true,
		},
		{
			name:    "optional sibling, none",
			in:      in2,
			primary: t1,
			join:    &schema.Join{Type: types.OptionalSibling, Source: "adres"},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			want: []map[string]any{
				{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen"},
			},
		},
		{
			name:    "optional parent/child, none",
			in:      in2,
			primary: t1,
			join:    &schema.Join{Type: types.OptionalParentChild, Source: "adres"},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			want: []map[string]any{
				{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen"},
			},
		},
		{
			name:    "forced sibling, none",
			in:      in2,
			primary: t1,
			join:    &schema.Join{Type: types.ForcedSibling, Source: "adres"},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			want: []map[string]any{
				{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen", "postcode": nil},
			},
		},
		{
			name:    "forced parent/child, none",
			in:      in2,
			primary: t1,
			join:    &schema.Join{Type: types.ForcedParentChild, Source: "adres"},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			want: []map[string]any{
				{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen", "adres": models.Rows{}},
			},
		},
		{
			name:    "optional sibling, one",
			in:      in3,
			primary: t1,
			join:    &schema.Join{Type: types.OptionalSibling, Source: "adres"},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			want: []map[string]any{
				{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen"},
				{"bsn": "999990408", "voornaam": "Piet", "achternaam": "Pietersen", "postcode": "2222YY"},
			},
		},
		{
			name:    "optional parent/child, one",
			in:      in3,
			primary: t1,
			join:    &schema.Join{Type: types.OptionalParentChild, Source: "adres"},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			want: []map[string]any{
				{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen"},
				{"bsn": "999990408", "voornaam": "Piet", "achternaam": "Pietersen", "adres": models.Rows{{Data: map[string]any{"postcode": "2222YY"}}}},
			},
		},
		{
			name:    "forced sibling, one",
			in:      in3,
			primary: t1,
			join:    &schema.Join{Type: types.ForcedSibling, Source: "adres"},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			want: []map[string]any{
				{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen", "postcode": nil},
				{"bsn": "999990408", "voornaam": "Piet", "achternaam": "Pietersen", "postcode": "2222YY"},
			},
		},
		{
			name:    "forced parent/child, one",
			in:      in3,
			primary: t1,
			join:    &schema.Join{Type: types.ForcedParentChild, Source: "adres"},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			want: []map[string]any{
				{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen", "adres": models.Rows{}},
				{"bsn": "999990408", "voornaam": "Piet", "achternaam": "Pietersen", "adres": models.Rows{{Data: map[string]any{"postcode": "2222YY"}}}},
			},
		},
		{
			name:    "optional sibling, few",
			in:      in1,
			primary: t1,
			join:    &schema.Join{Type: types.OptionalSibling, Source: "adres", QualifiedFields: true},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			want: []map[string]any{
				{"persoon.bsn": "999990391", "persoon.voornaam": "Jan", "persoon.achternaam": "Jansen", "adres.postcode": "1111ZZ"},
				{"persoon.bsn": "999990408", "persoon.voornaam": "Piet", "persoon.achternaam": "Pietersen", "adres.postcode": "2222YY"},
				{"persoon.bsn": "999990421", "persoon.voornaam": "Hendrik", "persoon.achternaam": "Hendriksen"},
			},
		},
		{
			name:    "optional parent/child, few",
			in:      in1,
			primary: t1,
			join:    &schema.Join{Type: types.OptionalParentChild, Source: "adres"},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			want: []map[string]any{
				{"bsn": "999990391", "voornaam": "Jan", "achternaam": "Jansen", "adres": models.Rows{{Data: map[string]any{"postcode": "1111ZZ"}}}},
				{"bsn": "999990408", "voornaam": "Piet", "achternaam": "Pietersen", "adres": models.Rows{{Data: map[string]any{"postcode": "2222YY"}}}},
				{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen"},
			},
		},
		{
			name:    "forced sibling, few",
			in:      in1,
			primary: t1,
			join:    &schema.Join{Type: types.ForcedSibling, Source: "adres", QualifiedFields: true},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			want: []map[string]any{
				{"persoon.bsn": "999990391", "persoon.voornaam": "Jan", "persoon.achternaam": "Jansen", "adres.postcode": "1111ZZ"},
				{"persoon.bsn": "999990408", "persoon.voornaam": "Piet", "persoon.achternaam": "Pietersen", "adres.postcode": "2222YY"},
				{"persoon.bsn": "999990421", "persoon.voornaam": "Hendrik", "persoon.achternaam": "Hendriksen", "adres.postcode": nil},
			},
		},
		{
			name:    "forced parent/child, few",
			in:      in1,
			primary: t1,
			join:    &schema.Join{Type: types.ForcedParentChild, Source: "adres"},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			want: []map[string]any{
				{"bsn": "999990391", "voornaam": "Jan", "achternaam": "Jansen", "adres": models.Rows{{Data: map[string]any{"postcode": "1111ZZ"}}}},
				{"bsn": "999990408", "voornaam": "Piet", "achternaam": "Pietersen", "adres": models.Rows{{Data: map[string]any{"postcode": "2222YY"}}}},
				{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen", "adres": models.Rows{}},
			},
		},
		{
			name:    "forced sibling, few, filtered one",
			in:      in1,
			primary: t1,
			join:    &schema.Join{Type: types.ForcedSibling, Source: "adres", QualifiedFields: true},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			filter: map[string]any{"postcode": "2222YY"},
			want: []map[string]any{
				{"persoon.bsn": "999990408", "persoon.voornaam": "Piet", "persoon.achternaam": "Pietersen", "adres.postcode": "2222YY"},
			},
		},
		{
			name:    "forced parent/child, few, filtered one",
			in:      in1,
			primary: t1,
			join:    &schema.Join{Type: types.ForcedParentChild, Source: "adres"},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			filter: map[string]any{"postcode": "1111ZZ"},
			want: []map[string]any{
				{"bsn": "999990391", "voornaam": "Jan", "achternaam": "Jansen", "adres": models.Rows{{Data: map[string]any{"postcode": "1111ZZ"}}}},
			},
		},
		{
			name:    "forced sibling, few, filtered none",
			in:      in1,
			primary: t1,
			join:    &schema.Join{Type: types.ForcedSibling, Source: "adres", QualifiedFields: true},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			filter: map[string]any{"postcode": "2222ZZ"},
			want:   []map[string]any{},
		},
		{
			name:    "forced parent/child, few, filtered none",
			in:      in1,
			primary: t1,
			join:    &schema.Join{Type: types.ForcedParentChild, Source: "adres"},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			filter: map[string]any{"postcode": "5555FF"},
			want:   []map[string]any{},
		},
		{
			name:    "optional sibling, few, filtered none",
			in:      in1,
			primary: t1,
			join:    &schema.Join{Type: types.OptionalSibling, Source: "adres", QualifiedFields: true},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			filter: map[string]any{"postcode": "2222ZZ"},
			want: []map[string]any{
				{"bsn": "999990391", "voornaam": "Jan", "achternaam": "Jansen"},
				{"bsn": "999990408", "voornaam": "Piet", "achternaam": "Pietersen"},
				{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen"},
			},
		},
		{
			name:    "optional parent/child, few, filtered none",
			in:      in1,
			primary: t1,
			join:    &schema.Join{Type: types.OptionalParentChild, Source: "adres"},
			fk: &schema.ForeignKey{
				Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
				ForeignTable: "persoon",
			},
			filter: map[string]any{"postcode": "5555FF"},
			want: []map[string]any{
				{"bsn": "999990391", "voornaam": "Jan", "achternaam": "Jansen"},
				{"bsn": "999990408", "voornaam": "Piet", "achternaam": "Pietersen"},
				{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			j := New(tc.primary, tc.in, tc.join, s)
			require.NotNil(t, j)

			source, err := s.GetTable(tc.join.Source)
			require.NoError(t, err)
			require.NotNil(t, source)

			tc.fk.Fix(source.Definition(), nil)

			got, err4 := j.JoinOnFK(tc.fk, tc.filter)
			if tc.wantErr {
				require.Error(t, err4)
				require.Nil(t, got)
			} else {
				require.NoError(t, err4)
				require.NotNil(t, got)
				require.Equal(t, len(tc.want), len(got))

				for i := range tc.want {
					m1, m2 := tc.want[i], got[i].Data
					assert.EqualExportedValues(t, m1, m2)
				}
			}
		})
	}
}

func TestJoin_JoinOnFields(t *testing.T) {
	t.Parallel()

	ds.Fix(nil)

	p1 := map[string]any{"bsn": "999990391", "voornaam": "Jan", "achternaam": "Jansen"}
	p2 := map[string]any{"bsn": "999990408", "voornaam": "Piet", "achternaam": "Pietersen"}
	p3 := map[string]any{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen"}

	t1 := models.TableFromData(persoon, []map[string]any{p1, p2, p3})

	t2 := models.TableFromData(adres, []map[string]any{
		{"bsn": "999990391", "postcode": "1111ZZ"},
		{"bsn": "999990408", "postcode": "2222YY"},
		{"bsn": "999990433", "postcode": "4444WW"},
	})

	s := &myMeta{tables: map[string]*models.Table{"persoon": t1, "adres": t2}}

	in1 := models.Rows{{Data: p1}, {Data: p2}, {Data: p3}}

	testCases := []struct {
		name    string
		in      models.Rows
		primary *models.Table
		join    *schema.Join
		fields  []string
		filter  map[string]any
		wantErr bool
		want    []map[string]any
	}{
		{
			name:    "forced parent/child, few, filtered one",
			in:      in1,
			primary: t1,
			join:    &schema.Join{Type: types.ForcedParentChild, Source: "adres"},
			fields:  []string{"bsn"},
			filter:  map[string]any{"postcode": "1111ZZ"},
			want: []map[string]any{
				{"bsn": "999990391", "voornaam": "Jan", "achternaam": "Jansen", "adres": models.Rows{{Data: map[string]any{"postcode": "1111ZZ"}}}},
			},
		},
		{
			name:    "forced sibling, few, filtered none",
			in:      in1,
			primary: t1,
			join:    &schema.Join{Type: types.ForcedSibling, Source: "adres", QualifiedFields: true},
			fields:  []string{"bsn"},
			filter:  map[string]any{"postcode": "2222ZZ"},
			want:    []map[string]any{},
		},
		{
			name:    "forced parent/child, few, filtered none",
			in:      in1,
			primary: t1,
			join:    &schema.Join{Type: types.ForcedParentChild, Source: "adres"},
			fields:  []string{"bsn"},
			filter:  map[string]any{"postcode": "5555FF"},
			want:    []map[string]any{},
		},
		{
			name:    "optional sibling, few, filtered none",
			in:      in1,
			primary: t1,
			join:    &schema.Join{Type: types.OptionalSibling, Source: "adres", QualifiedFields: true},
			fields:  []string{"bsn"},
			filter:  map[string]any{"postcode": "2222ZZ"},
			want: []map[string]any{
				{"bsn": "999990391", "voornaam": "Jan", "achternaam": "Jansen"},
				{"bsn": "999990408", "voornaam": "Piet", "achternaam": "Pietersen"},
				{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen"},
			},
		},
		{
			name:    "optional parent/child, few, filtered none",
			in:      in1,
			primary: t1,
			join:    &schema.Join{Type: types.OptionalParentChild, Source: "adres"},
			fields:  []string{"bsn"},
			filter:  map[string]any{"postcode": "5555FF"},
			want: []map[string]any{
				{"bsn": "999990391", "voornaam": "Jan", "achternaam": "Jansen"},
				{"bsn": "999990408", "voornaam": "Piet", "achternaam": "Pietersen"},
				{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			j := New(tc.primary, tc.in, tc.join, s)
			require.NotNil(t, j)

			got, err4 := j.JoinOnFields(tc.fields, tc.filter)
			if tc.wantErr {
				require.Error(t, err4)
				require.Nil(t, got)
			} else {
				require.NoError(t, err4)
				require.NotNil(t, got)
				require.Equal(t, len(tc.want), len(got))

				for i := range tc.want {
					m1, m2 := tc.want[i], got[i].Data
					assert.EqualExportedValues(t, m1, m2)
				}
			}
		})
	}
}

func TestJoin_SubJoin(t *testing.T) {
	t.Parallel()

	ds.Fix(nil)

	p1 := map[string]any{"bsn": "999990391", "voornaam": "Jan", "achternaam": "Jansen"}
	p2 := map[string]any{"bsn": "999990408", "voornaam": "Piet", "achternaam": "Pietersen"}
	p3 := map[string]any{"bsn": "999990421", "voornaam": "Hendrik", "achternaam": "Hendriksen"}

	t1 := models.TableFromData(persoon, []map[string]any{p1, p2, p3})

	t2 := models.TableFromData(adres, []map[string]any{
		{"bsn": "999990391", "postcode": "1111ZZ"},
		{"bsn": "999990408", "postcode": "2222YY"},
		{"bsn": "999990433", "postcode": "4444WW"},
	})

	t3 := models.TableFromData(inwoners, []map[string]any{
		{"postcode": "1111ZZ", "inwoners": 1},
		{"postcode": "2222YY", "inwoners": 7},
		{"postcode": "4444WW", "inwoners": 3},
	})

	s := &myMeta{tables: map[string]*models.Table{"persoon": t1, "adres": t2, "inwoners": t3}}

	in1 := models.Rows{{Data: p1}, {Data: p2}, {Data: p3}}

	t.Run("sub-join", func(t *testing.T) {
		t.Parallel()

		j1 := New(t1, in1, &schema.Join{Type: types.OptionalSibling, Source: "adres"}, s)
		require.NotNil(t, j1)

		got1, err3 := j1.JoinOnFK(t2.Definition().ForeignKey("fk1"), nil)
		require.NoError(t, err3)
		require.NotNil(t, got1)
		assert.Len(t, got1, 3)

		j2 := New(t1, got1, &schema.Join{Type: types.OptionalSibling, Source: "inwoners"}, s)
		require.NotNil(t, j1)

		got2, err4 := j2.JoinOnFields([]string{"postcode"}, nil)
		require.NoError(t, err4)
		require.NotNil(t, got2)
		assert.Len(t, got2, 2)
	})

}

var persoon = &schema.Table{
	Object: schema.Object{
		Parent:      schema.Parent{ID: "persoon"},
		Description: "personen",
		Fields: []*schema.Field{
			{Object: schema.Object{Parent: schema.Parent{ID: "bsn"}, Description: "Burgerservicenummer"}, IsPII: true},
			{Object: schema.Object{Parent: schema.Parent{ID: "voornaam"}, Description: "Voornaam"}, IsPII: true},
			{Object: schema.Object{Parent: schema.Parent{ID: "achternaam"}, Description: "Achternaam"}, IsPII: true},
		},
	},
	PrimaryKey: &schema.Index{Parent: schema.Parent{ID: "pk"}, Description: "primary key", Fields: []string{"bsn"}},
}

var adres = &schema.Table{
	Object: schema.Object{
		Parent:      schema.Parent{ID: "adres"},
		Description: "adressen",
		Fields: []*schema.Field{
			{Object: schema.Object{Parent: schema.Parent{ID: "bsn"}, Description: "Burgerservicenummer"}, IsPII: true},
			{Object: schema.Object{Parent: schema.Parent{ID: "postcode"}, Description: "Postcode"}},
		},
	},
	PrimaryKey: &schema.Index{Parent: schema.Parent{ID: "pk"}, Description: "primary key", Fields: []string{"bsn"}},
	ForeignKeys: []*schema.ForeignKey{
		{
			Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"bsn"}},
			ForeignTable: "persoon",
		},
	},
}

var inwoners = &schema.Table{
	Object: schema.Object{
		Parent:      schema.Parent{ID: "inwoner"},
		Description: "inwoners",
		Fields: []*schema.Field{
			{Object: schema.Object{Parent: schema.Parent{ID: "postcode"}, Description: "Postcode"}},
			{Object: schema.Object{Parent: schema.Parent{ID: "inwoners"}, Description: "Aantal inwoners"}},
		},
	},
	PrimaryKey: &schema.Index{Parent: schema.Parent{ID: "pk"}, Description: "primary key", Fields: []string{"postcode"}},
	ForeignKeys: []*schema.ForeignKey{
		{
			Index:        schema.Index{Parent: schema.Parent{ID: "fk1"}, Fields: []string{"postcode"}},
			ForeignTable: "adres",
		},
	},
}

var kenteken = &schema.Table{
	Object: schema.Object{
		Parent:      schema.Parent{ID: "kenteken"},
		Description: "kentekens",
		Fields: []*schema.Field{
			{Object: schema.Object{Parent: schema.Parent{ID: "kenteken"}, Description: "kenteken"}, IsPII: true},
		},
	},
	PrimaryKey: &schema.Index{Parent: schema.Parent{ID: "pk"}, Description: "primary key", Fields: []string{"kenteken"}},
}

var ds = &schema.Datasource{
	Parent: schema.Parent{ID: "ds"},
	Tables: []*schema.Table{persoon, adres, kenteken, inwoners},
}

type myMeta struct {
	tables map[string]*models.Table
}

func (s *myMeta) IterateEndpoints(_ func(def *schema.Endpoint)) {}
func (s *myMeta) DatasourceExists(_ string) bool                { return true }
func (s *myMeta) TableExists(_ string) bool                     { return true }
func (s *myMeta) GetDataspace() *models.Dataspace               { return nil }
func (s *myMeta) GetDatasources() map[string]*models.Datasource { return nil }
func (s *myMeta) GetDatasource(_ string) *models.Datasource     { return nil }
func (s *myMeta) GetTable(id string) (*models.Table, error)     { return s.tables[id], nil }
