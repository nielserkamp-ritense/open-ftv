package filtering

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
)

func TestFilterFromQuery(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		primary string
		q       map[string]string
		want    *Filter
	}{
		{
			name: "empty",
			q:    nil,
			want: &Filter{AllOfFilter: AllOfFilter{AllOf: make(Filters, 0)}},
		},
		{
			name: "single field",
			q:    map[string]string{"f1": "hello world"},
			want: &Filter{AllOfFilter: AllOfFilter{AllOf: Filters{
				&Filter{FieldValueFilter: FieldValueFilter{
					Level:   enums.PrimaryLevel,
					Field:   "f1",
					Compare: enums.IsEqual,
					Value:   "hello world",
				}},
			}}},
		},
		{
			name:    "few fields",
			primary: "t1",
			q:       map[string]string{"f1": "hello world", "t1.f2": "123", "j1.f3": "true"},
			want: &Filter{AllOfFilter: AllOfFilter{AllOf: Filters{
				&Filter{FieldValueFilter: FieldValueFilter{
					Level:   enums.PrimaryLevel,
					Field:   "f1",
					Compare: enums.IsEqual,
					Value:   "hello world",
				}},
				&Filter{FieldValueFilter: FieldValueFilter{
					Level:   enums.JoinLevel,
					Join:    "j1",
					Field:   "f3",
					Compare: enums.IsEqual,
					Value:   "true",
				}},
				&Filter{FieldValueFilter: FieldValueFilter{
					Level:   enums.PrimaryLevel,
					Table:   "t1",
					Field:   "f2",
					Compare: enums.IsEqual,
					Value:   "123",
				}},
			}}},
		},
		{
			name: "filter field",
			q:    map[string]string{"@filter": `{"level": "primary", "field": "f1", "compare": "greater", "value": 80000}`},
			want: &Filter{AllOfFilter: AllOfFilter{AllOf: Filters{
				&Filter{FieldValueFilter: FieldValueFilter{
					Level:   enums.PrimaryLevel,
					Field:   "f1",
					Compare: enums.IsGreater,
					Value:   float64(80000),
				}},
			}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FilterFromQuery(tc.primary, tc.q)
			require.NotNil(t, got)

			if len(tc.want.AllOf) == 1 {
				require.Nil(t, got.AllOf)
				require.Nil(t, got.AnyOf)
				assert.EqualValues(t, tc.want.AllOf[0], got)
			} else {
				require.NotNil(t, got.AllOf)
				require.Nil(t, got.AnyOf)
				require.Equal(t, len(tc.want.AllOf), len(got.AllOf))

				for i, filter1 := range tc.want.AllOf {
					assert.EqualValues(t, filter1, got.AllOf[i])
				}
			}
		})
	}
}

func TestFilterFromMap(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		primary string
		q       map[string]any
		want    *Filter
	}{
		{
			name: "empty",
			q:    nil,
			want: &Filter{AllOfFilter: AllOfFilter{AllOf: make(Filters, 0)}},
		},
		{
			name: "single field",
			q:    map[string]any{"f1": "hello world"},
			want: &Filter{AllOfFilter: AllOfFilter{AllOf: Filters{
				&Filter{FieldValueFilter: FieldValueFilter{
					Level:   enums.PrimaryLevel,
					Field:   "f1",
					Compare: enums.IsEqual,
					Value:   "hello world",
				}},
			}}},
		},
		{
			name:    "few fields",
			primary: "t1",
			q:       map[string]any{"f1": "hello world", "t1.f2": 123, "j1.f3": true},
			want: &Filter{AllOfFilter: AllOfFilter{AllOf: Filters{
				&Filter{FieldValueFilter: FieldValueFilter{
					Level:   enums.PrimaryLevel,
					Field:   "f1",
					Compare: enums.IsEqual,
					Value:   "hello world",
				}},
				&Filter{FieldValueFilter: FieldValueFilter{
					Level:   enums.JoinLevel,
					Join:    "j1",
					Field:   "f3",
					Compare: enums.IsEqual,
					Value:   true,
				}},
				&Filter{FieldValueFilter: FieldValueFilter{
					Level:   enums.PrimaryLevel,
					Table:   "t1",
					Field:   "f2",
					Compare: enums.IsEqual,
					Value:   123,
				}},
			}}},
		},
		{
			name: "filter field",
			q:    map[string]any{"@filter": `{"level": "primary", "field": "f1", "compare": "greater", "value": 80000}`},
			want: &Filter{AllOfFilter: AllOfFilter{AllOf: Filters{
				&Filter{FieldValueFilter: FieldValueFilter{
					Level:   enums.PrimaryLevel,
					Field:   "f1",
					Compare: enums.IsGreater,
					Value:   float64(80000),
				}},
			}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FilterFromMap(tc.primary, tc.q)
			require.NotNil(t, got)

			if len(tc.want.AllOf) == 1 {
				require.Nil(t, got.AllOf)
				require.Nil(t, got.AnyOf)
				assert.EqualValues(t, tc.want.AllOf[0], got)
			} else {
				require.NotNil(t, got.AllOf)
				require.Nil(t, got.AnyOf)
				require.Equal(t, len(tc.want.AllOf), len(got.AllOf))

				for i, filter1 := range tc.want.AllOf {
					assert.EqualValues(t, filter1, got.AllOf[i])
				}
			}
		})
	}
}

func TestFilterFromString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		primary string
		in      string
		want    *Filter
	}{
		{
			name: "empty",
			want: &Filter{AllOfFilter: AllOfFilter{AllOf: make(Filters, 0)}},
		},
		{
			name:    "single compare",
			primary: "t1",
			in:      "postcode=1111ZZ",
			want: &Filter{AllOfFilter: AllOfFilter{AllOf: []*Filter{
				{FieldValueFilter: FieldValueFilter{Level: enums.PrimaryLevel, Field: "postcode", Compare: enums.IsEqual, Value: "1111ZZ"}},
			}}},
		},
		{
			name:    "few compares",
			primary: "t1",
			in:      "adres.postcode <= 1111ZZ, volwassen != false, bsn NOT NIL",
			want: &Filter{AllOfFilter: AllOfFilter{AllOf: []*Filter{
				{FieldValueFilter: FieldValueFilter{Level: enums.JoinLevel, Join: "adres", Field: "postcode", Compare: enums.IsLesserOrEqual, Value: "1111ZZ"}},
				{FieldValueFilter: FieldValueFilter{Level: enums.PrimaryLevel, Field: "volwassen", Compare: enums.IsNotEqual, Value: "false"}},
				{FieldValueFilter: FieldValueFilter{Level: enums.PrimaryLevel, Field: "bsn", Compare: enums.Exists}},
			}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FilterFromString(tc.primary, tc.in)
			require.NotNil(t, got)
		})
	}
}
