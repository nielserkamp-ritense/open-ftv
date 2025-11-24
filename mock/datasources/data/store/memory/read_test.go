package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/context"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/matching"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/store/memory/models"
)

func TestStorage_SelectPK(t *testing.T) {
	s1 := New(mockFDS)
	require.NotNil(t, s1)

	loadFDS(s1)

	m1 := matching.NewFieldMatcher("voornaam")

	testCases := []struct {
		name    string
		s       store.Storage
		table   string
		pk      []any
		matcher matching.FieldMatcher
		wantErr bool
		want    *models.Row
	}{
		{
			name:    "bad datasource",
			s:       s1,
			table:   "xyz.persoon",
			wantErr: true,
		},
		{
			name:    "bad table",
			s:       s1,
			table:   "brp.xyz",
			wantErr: true,
		},
		{
			name:    "bad parameter count",
			s:       s1,
			table:   "persoon",
			pk:      []any{1, 2, 3},
			wantErr: true,
		},
		{
			name:    "not found",
			s:       s1,
			table:   "persoon",
			pk:      []any{"xyz"},
			wantErr: true,
		},
		{
			name:  "found - no filter",
			s:     s1,
			table: "persoon",
			pk:    []any{"999990263"},
			want:  &models.Row{Data: map[string]any{"bsn": "999990263", "voornaam": "Gerrit", "achternaam": "Bommels"}},
		},
		{
			name:    "found - with filter",
			s:       s1,
			table:   "persoon",
			pk:      []any{"999990263"},
			matcher: m1,
			want:    &models.Row{Data: map[string]any{"voornaam": "Gerrit"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, modified, err2 := tc.s.SelectPK(tc.table, tc.pk, tc.matcher)
			if tc.wantErr {
				require.Error(t, err2)
				require.Nil(t, got)
				require.Nil(t, modified)
			} else {
				require.NoError(t, err2)
				require.NotNil(t, got)
				require.NotNil(t, modified)
				assert.EqualValues(t, tc.want.Data, got.Data)
			}
		})
	}
}

func TestStorage_SelectIX(t *testing.T) {
	s1 := New(mockFDS)
	require.NotNil(t, s1)

	loadFDS(s1)

	m1 := matching.NewFieldMatcher("voornaam")

	testCases := []struct {
		name    string
		s       store.Storage
		table   string
		index   string
		keys    []any
		matcher matching.FieldMatcher
		wantErr bool
		want    models.Rows
	}{
		{
			name:    "bad datasource",
			s:       s1,
			table:   "xyz.persoon",
			wantErr: true,
		},
		{
			name:    "bad table",
			s:       s1,
			table:   "brp.xyz",
			wantErr: true,
		},
		{
			name:    "no index",
			s:       s1,
			table:   "brp.persoon",
			wantErr: true,
		},
		{
			name:    "bad index",
			s:       s1,
			table:   "brp.persoon",
			index:   "xyz",
			wantErr: true,
		},
		{
			name:    "bad parameter count",
			s:       s1,
			table:   "persoon",
			index:   "ix2",
			keys:    []any{1, 2, 3},
			wantErr: true,
		},
		{
			name:    "not found",
			s:       s1,
			table:   "persoon",
			index:   "ix2",
			keys:    []any{"xyz"},
			wantErr: true,
		},
		{
			name:  "found - no filter",
			s:     s1,
			table: "persoon",
			index: "ix2",
			keys:  []any{"Gerrit"},
			want:  models.Rows{{Data: map[string]any{"bsn": "999990263", "voornaam": "Gerrit", "achternaam": "Bommels"}}},
		},
		{
			name:    "found - with filter",
			s:       s1,
			table:   "persoon",
			index:   "ix2",
			keys:    []any{"Gerrit"},
			matcher: m1,
			want:    models.Rows{{Data: map[string]any{"voornaam": "Gerrit"}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, modified, err2 := tc.s.SelectIX(tc.table, tc.index, tc.keys, tc.matcher)
			if tc.wantErr {
				require.Error(t, err2)
				require.Nil(t, got)
				require.Nil(t, modified)
			} else {
				require.NoError(t, err2)
				require.NotNil(t, got)
				require.NotNil(t, modified)
				require.Equal(t, len(tc.want), len(got))

				for i := range tc.want {
					assert.EqualValues(t, tc.want[i].Data, got[i].Data)
				}
			}
		})
	}
}

func TestStorage_Search(t *testing.T) {
	s1 := New(mockFDS)
	require.NotNil(t, s1)

	loadFDS(s1)

	ctx0, err0 := context.New(map[string]string{"@filter": "persoon.bsn = xyz"}, nil, "persoon")
	require.NoError(t, err0)
	err0 = ctx0.Filter.Prepare(ds1, nil)
	require.NoError(t, err0)

	ctx1, err := context.New(nil, nil, "persoon")
	require.NoError(t, err)
	err = ctx1.Filter.Prepare(ds1, nil)
	require.NoError(t, err)

	ctx2, err2 := context.New(map[string]string{"@filter": "persoon.bsn > 999990270"}, nil, "persoon")
	require.NoError(t, err2)
	err2 = ctx2.Filter.Prepare(ds1, nil)
	require.NoError(t, err2)

	ctx3, err3 := context.New(map[string]string{"@filter": "voornaam like %sula"}, nil, "persoon")
	require.NoError(t, err3)
	err3 = ctx3.Filter.Prepare(ds1, nil)
	require.NoError(t, err3)

	ctx4, err4 := context.New(map[string]string{"@filter": "voornaam like %sula", "@fields": "voornaam"}, nil, "persoon")
	require.NoError(t, err4)
	err4 = ctx4.Filter.Prepare(ds1, nil)
	require.NoError(t, err4)

	testCases := []struct {
		name    string
		s       store.Storage
		table   string
		ctx     *context.RequestContext
		wantErr bool
		want    models.Rows
	}{
		{
			name:    "bad datasource",
			s:       s1,
			table:   "xyz.persoon",
			wantErr: true,
		},
		{
			name:    "bad table",
			s:       s1,
			table:   "brp.xyz",
			wantErr: true,
		},
		{
			name:    "not found",
			s:       s1,
			table:   "persoon",
			ctx:     ctx0,
			wantErr: true,
		},
		{
			name:  "found all",
			s:     s1,
			table: "persoon",
			ctx:   ctx1,
			want: models.Rows{
				{Data: map[string]any{"bsn": "999990251", "voornaam": "Pieter Koen", "achternaam": "Wezichem"}},
				{Data: map[string]any{"bsn": "999990263", "voornaam": "Gerrit", "achternaam": "Bommels"}},
				{Data: map[string]any{"bsn": "999990275", "voornaam": "Amaria", "achternaam": "Bont"}},
				{Data: map[string]any{"bsn": "999990287", "voornaam": "Ursula", "achternaam": "Koenders-van Zanten"}},
			},
		},
		{
			name:  "found subset",
			s:     s1,
			table: "persoon",
			ctx:   ctx2,
			want: models.Rows{
				{Data: map[string]any{"bsn": "999990275", "voornaam": "Amaria", "achternaam": "Bont"}},
				{Data: map[string]any{"bsn": "999990287", "voornaam": "Ursula", "achternaam": "Koenders-van Zanten"}},
			},
		},
		{
			name:  "found exact - no filter",
			s:     s1,
			table: "persoon",
			ctx:   ctx3,
			want: models.Rows{
				{Data: map[string]any{"bsn": "999990287", "voornaam": "Ursula", "achternaam": "Koenders-van Zanten"}},
			},
		},
		{
			name:  "found exact - with filter",
			s:     s1,
			table: "persoon",
			ctx:   ctx4,
			want: models.Rows{
				{Data: map[string]any{"voornaam": "Ursula"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, modified, err9 := tc.s.Search(tc.table, tc.ctx)
			if tc.wantErr {
				require.Error(t, err9)
				require.Nil(t, got)
				require.Nil(t, modified)
			} else {
				require.NoError(t, err9)
				require.NotNil(t, got)
				require.NotNil(t, modified)
				require.Equal(t, len(tc.want), len(got))

				for i := range tc.want {
					assert.EqualValues(t, tc.want[i].Data, got[i].Data)
				}
			}
		})
	}
}

func TestStorage_GetEndpoint(t *testing.T) {
	s1 := New(mockFDS)
	require.NotNil(t, s1)

	loadFDS(s1)

	e1 := &schema.Endpoint{
		Version:     1,
		Type:        enums.GetMethod,
		CalledAs:    enums.GetMethod,
		Path:        "/path",
		FullVersion: "1.0.0",
		Datasource:  "xyz",
		Table:       "persoon",
	}

	e2 := &schema.Endpoint{
		Version:     1,
		Type:        enums.GetMethod,
		CalledAs:    enums.GetMethod,
		Path:        "/path",
		FullVersion: "1.0.0",
		Datasource:  "brp",
		Table:       "xyz",
	}

	e3 := &schema.Endpoint{
		Version:     1,
		Type:        enums.GetMethod,
		CalledAs:    enums.PostMethod,
		Path:        "/path",
		FullVersion: "1.0.0",
		Datasource:  "brp",
		Table:       "persoon",
	}

	e4 := &schema.Endpoint{
		Version:     1,
		Type:        enums.GetMethod,
		CalledAs:    enums.GetMethod,
		Path:        "/path",
		FullVersion: "1.0.0",
		Datasource:  "brp",
		Table:       "persoon",
	}

	s1.AddEndpoint(e1)
	s1.AddEndpoint(e2)
	s1.AddEndpoint(e3)
	s1.AddEndpoint(e4)

	ctx0, err0 := context.New(map[string]string{"@filter": "persoon.bsn = xyz"}, nil, "persoon")
	require.NoError(t, err0)
	err0 = ctx0.Filter.Prepare(ds1, nil)
	require.NoError(t, err0)

	ctx1, err := context.New(nil, nil, "persoon")
	require.NoError(t, err)
	err = ctx1.Filter.Prepare(ds1, nil)
	require.NoError(t, err)

	ctx2, err2 := context.New(map[string]string{"@filter": "persoon.bsn > 999990270"}, nil, "persoon")
	require.NoError(t, err2)
	err2 = ctx2.Filter.Prepare(ds1, nil)
	require.NoError(t, err2)

	ctx3, err3 := context.New(map[string]string{"@filter": "voornaam like %sula"}, nil, "persoon")
	require.NoError(t, err3)
	err3 = ctx3.Filter.Prepare(ds1, nil)
	require.NoError(t, err3)

	ctx4, err4 := context.New(map[string]string{"@filter": "voornaam like %sula", "@fields": "voornaam"}, nil, "persoon")
	require.NoError(t, err4)
	err4 = ctx4.Filter.Prepare(ds1, nil)
	require.NoError(t, err4)

	testCases := []struct {
		name    string
		s       store.Storage
		e       *schema.Endpoint
		ctx     *context.RequestContext
		wantErr bool
		want    models.Rows
	}{
		{
			name:    "bad datasource",
			s:       s1,
			e:       e1,
			wantErr: true,
		},
		{
			name:    "bad table",
			s:       s1,
			e:       e2,
			wantErr: true,
		},
		{
			name:    "not found",
			s:       s1,
			e:       e3,
			ctx:     ctx0,
			wantErr: false,
			want:    nil,
		},
		{
			name: "found all",
			s:    s1,
			e:    e4,
			ctx:  ctx1,
			want: models.Rows{
				{Data: map[string]any{"bsn": "999990251", "voornaam": "Pieter Koen", "achternaam": "Wezichem"}},
				{Data: map[string]any{"bsn": "999990263", "voornaam": "Gerrit", "achternaam": "Bommels"}},
				{Data: map[string]any{"bsn": "999990275", "voornaam": "Amaria", "achternaam": "Bont"}},
				{Data: map[string]any{"bsn": "999990287", "voornaam": "Ursula", "achternaam": "Koenders-van Zanten"}},
			},
		},
		{
			name: "found subset",
			s:    s1,
			e:    e4,
			ctx:  ctx2,
			want: models.Rows{
				{Data: map[string]any{"bsn": "999990275", "voornaam": "Amaria", "achternaam": "Bont"}},
				{Data: map[string]any{"bsn": "999990287", "voornaam": "Ursula", "achternaam": "Koenders-van Zanten"}},
			},
		},
		{
			name: "found exact - no filter",
			s:    s1,
			e:    e4,
			ctx:  ctx3,
			want: models.Rows{
				{Data: map[string]any{"bsn": "999990287", "voornaam": "Ursula", "achternaam": "Koenders-van Zanten"}},
			},
		},
		{
			name: "found exact - with filter",
			s:    s1,
			e:    e4,
			ctx:  ctx4,
			want: models.Rows{
				{Data: map[string]any{"voornaam": "Ursula"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, modified, err9 := tc.s.GetEndpoint(tc.e, tc.ctx)
			if tc.wantErr {
				require.Error(t, err9)
				require.Nil(t, got)
				require.Nil(t, modified)
			} else {
				require.NoError(t, err9)
				require.Equal(t, len(tc.want), len(got))

				for i := range tc.want {
					assert.EqualValues(t, tc.want[i].Data, got[i].Data)
				}
			}
		})
	}
}
