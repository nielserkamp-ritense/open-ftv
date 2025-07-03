package filtering

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
)

func TestFilter_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		data      string
		wantErr   bool
		wantField bool
		wantAll   bool
		wantAny   bool
		want      string
	}{
		{
			name:    "bad data",
			data:    "\001\002",
			wantErr: true,
		},
		{
			name:      "field value",
			data:      `{"level":"primary"}`,
			wantField: true,
			want:      "{level=primary,field=,compare-type=[unknown]}",
		},
		{
			name:    "all of",
			data:    `{"allOf":[{"level":"primary"},{"level":"primary"}]}`,
			wantAll: true,
			want:    "{allOf:[{level=primary,field=,compare-type=[unknown]} {level=primary,field=,compare-type=[unknown]}]}",
		},
		{
			name:    "any of",
			data:    `{"anyOf":[{"level":"primary"},{"level":"primary"}]}`,
			wantAny: true,
			want:    "{anyOf:[{level=primary,field=,compare-type=[unknown]} {level=primary,field=,compare-type=[unknown]}]}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := new(Filter)
			err := f.UnmarshalJSON([]byte(tc.data))
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.False(t, f.AllOf != nil && f.AnyOf != nil)
				assert.Equal(t, tc.wantField, f.AllOf == nil && f.AnyOf == nil)
				assert.Equal(t, tc.wantAll, f.AllOf != nil && f.AnyOf == nil)
				assert.Equal(t, tc.wantAny, f.AllOf == nil && f.AnyOf != nil)
				assert.Equal(t, tc.want, f.String())

				err = f.Prepare(nil, nil)
				require.Error(t, err)
			}
		})
	}
}

func TestFilter_UnmarshalYAML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		data      string
		wantErr   bool
		wantField bool
		wantAll   bool
		wantAny   bool
		want      string
	}{
		{
			name:    "bad data",
			data:    "\001\002",
			wantErr: true,
		},
		{
			name:      "field value",
			data:      `level: primary`,
			wantField: true,
			want:      "{level=primary,field=,compare-type=[unknown]}",
		},
		{
			name: "all of",
			data: `allOf:
  - level: primary
  - level: primary
`,
			wantAll: true,
			want:    "{allOf:[{level=primary,field=,compare-type=[unknown]} {level=primary,field=,compare-type=[unknown]}]}",
		},
		{
			name: "any of",
			data: `anyOf:
  - level: primary
  - level: primary
`,
			wantAny: true,
			want:    "{anyOf:[{level=primary,field=,compare-type=[unknown]} {level=primary,field=,compare-type=[unknown]}]}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := new(Filter)
			err := f.UnmarshalYAML([]byte(tc.data))
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.False(t, f.AllOf != nil && f.AnyOf != nil)
				assert.Equal(t, tc.wantField, f.AllOf == nil && f.AnyOf == nil)
				assert.Equal(t, tc.wantAll, f.AllOf != nil && f.AnyOf == nil)
				assert.Equal(t, tc.wantAny, f.AllOf == nil && f.AnyOf != nil)
				assert.Equal(t, tc.want, f.String())

				err = f.Prepare(nil, nil)
				require.Error(t, err)
			}
		})
	}
}

func TestFilterFromAny(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		primary string
		in      string
		want    *Filter
	}{
		{
			name: "json",
			in:   `{"level":"primary","caseInsensitive":true,"field":"leeftijd","compare":"IsGreater","value":17}`,
			want: &Filter{
				FieldValueFilter: FieldValueFilter{
					Level:       enums.PrimaryLevel,
					Insensitive: true,
					Field:       "leeftijd",
					Compare:     enums.IsGreater,
					Value:       float64(17),
				},
			},
		},
		{
			name: "yaml",
			in: `---
level: primary
caseInsensitive: true
field: leeftijd
compare: IsGreater
value: 17
`,
			want: &Filter{
				FieldValueFilter: FieldValueFilter{
					Level:       enums.PrimaryLevel,
					Insensitive: true,
					Field:       "leeftijd",
					Compare:     enums.IsGreater,
					Value:       uint64(17),
				},
			},
		},
		{
			name:    "primary table",
			primary: "persoon",
			in:      `persoon.leeftijd >= 18`,
			want: &Filter{
				FieldValueFilter: FieldValueFilter{
					Level:   enums.PrimaryLevel,
					Table:   "persoon",
					Field:   "leeftijd",
					Compare: enums.IsGreaterOrEqual,
					Value:   "18",
				},
			},
		},
		{
			name:    "join",
			primary: "persoon",
			in:      `adres.postcode not nil`,
			want: &Filter{
				FieldValueFilter: FieldValueFilter{
					Level:   enums.JoinLevel,
					Join:    "adres",
					Field:   "postcode",
					Compare: enums.Exists,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FilterFromAny(tc.primary, tc.in)
			assert.EqualValues(t, tc.want, got)
		})
	}
}
