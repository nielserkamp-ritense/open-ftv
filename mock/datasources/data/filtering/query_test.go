package filtering

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

func TestFilterFromQuery(t *testing.T) {
	t.Parallel()

	t1 := &schema.Table{Object: schema.Object{Parent: schema.Parent{ID: "t1"}}}
	t2 := &schema.Table{Object: schema.Object{Parent: schema.Parent{ID: "t2"}}}

	ds := &schema.Datasource{Parent: schema.Parent{ID: "ds"}, Tables: []*schema.Table{t1, t2}}
	ds.Fix(nil)

	testCases := []struct {
		name    string
		ds      *schema.Datasource
		primary string
		q       map[string]string
		want    *Filter
	}{
		{
			name: "empty",
			ds:   ds,
			q:    nil,
			want: &Filter{AllOfFilter: AllOfFilter{AllOf: make(Filters, 0)}},
		},
		{
			name: "single field",
			ds:   ds,
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
			ds:      ds,
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
					Level:   enums.AnyLevel,
					Table:   "t1",
					Field:   "f2",
					Compare: enums.IsEqual,
					Value:   "123",
				}},
			}}},
		},
		{
			name: "filter field",
			ds:   ds,
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

			got := FilterFromQuery(tc.ds, tc.primary, tc.q)
			require.NotNil(t, got)

			got2, ok := got.(*Filter)
			require.True(t, ok)
			require.NotNil(t, got2)
			require.NotNil(t, got2.AllOf)
			require.Nil(t, got2.AnyOf)
			require.Equal(t, len(tc.want.AllOf), len(got2.AllOf))

			for i, filter1 := range tc.want.AllOf {
				filter2 := got2.AllOf[i]
				assert.EqualValues(t, filter1, filter2)
			}
		})
	}
}
