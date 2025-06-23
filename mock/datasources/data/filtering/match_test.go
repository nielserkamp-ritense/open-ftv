package filtering

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

func TestFieldValueFilter_MatchOnPrimaryData(t *testing.T) {
	t.Parallel()

	f1 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: enums.StringType}

	tr1 := &schema.Transformation{
		Object:             schema.Object{Parent: schema.Parent{ID: "tr1"}},
		TransformationType: enums.TransformAge,
		ResultType:         enums.IntegerType,
		IsPII:              true,
		InputFields:        map[int]string{1: "f1"},
	}

	t1 := &schema.Table{
		Object: schema.Object{
			Parent: schema.Parent{ID: "t1"},
			Fields: []*schema.Field{f1},
		},
		Transforms: []*schema.Transformation{tr1},
	}

	ds := &schema.Datasource{
		Parent: schema.Parent{ID: "ds"},
		Tables: []*schema.Table{t1},
	}
	ds.Fix(nil)

	testCases := []struct {
		name string
		f    *FieldValueFilter
		data map[string]any
		want bool
	}{
		{
			name: "join level",
			f:    &FieldValueFilter{Level: enums.JoinLevel, field: f1, Compare: enums.Exists},
			data: map[string]any{},
			want: true,
		},
		{
			name: "exists - not found",
			f:    &FieldValueFilter{Level: enums.PrimaryLevel, field: f1, Compare: enums.Exists},
			data: map[string]any{},
		},
		{
			name: "field exists - found",
			f:    &FieldValueFilter{Level: enums.AnyLevel, field: f1, Compare: enums.Exists},
			data: map[string]any{"f1": "hello world"},
			want: true,
		},
		{
			name: "transform exists - found",
			f:    &FieldValueFilter{Level: enums.AnyLevel, transform: tr1, Compare: enums.Exists},
			data: map[string]any{"tr1": "hello world"},
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.f.MatchOnPrimaryData(tc.data)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFieldValueFilter_MatchOnJoinData(t *testing.T) {
	t.Parallel()

	f1 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: enums.StringType}
	t1 := &schema.Table{Object: schema.Object{Parent: schema.Parent{ID: "t1"}, Fields: []*schema.Field{f1}}}
	t2 := &schema.Table{Object: schema.Object{Parent: schema.Parent{ID: "t2"}, Fields: []*schema.Field{f1}}}

	ds := &schema.Datasource{Parent: schema.Parent{ID: "ds"}, Tables: []*schema.Table{t1, t2}}
	ds.Fix(nil)

	j1 := &schema.Join{Source: "t1", JoinID: "j1"}
	j1.Fix(ds)

	j2 := &schema.Join{Source: "t2", JoinID: "j2"}
	j2.Fix(ds)

	testCases := []struct {
		name string
		f    *FieldValueFilter
		j    *schema.Join
		data map[string]any
		want bool
	}{
		{
			name: "primary level",
			f:    &FieldValueFilter{Level: enums.PrimaryLevel, field: f1, Compare: enums.Exists},
			want: true,
		},
		{
			name: "different join",
			f:    &FieldValueFilter{Level: enums.JoinLevel, join: j1, field: f1, Compare: enums.Exists},
			j:    j2,
			want: true,
		},
		{
			name: "different table",
			f:    &FieldValueFilter{Level: enums.AnyLevel, join: j2, table: t1, field: f1, Compare: enums.Exists},
			j:    j2,
			want: true,
		},
		{
			name: "exists - not found",
			f:    &FieldValueFilter{Level: enums.JoinLevel, join: j1, table: t1, field: f1, Compare: enums.Exists},
			j:    j1,
			data: map[string]any{},
		},
		{
			name: "exists - found",
			f:    &FieldValueFilter{Level: enums.AnyLevel, join: j1, table: t1, field: f1, Compare: enums.Exists},
			j:    j1,
			data: map[string]any{"f1": "hello world"},
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.f.MatchOnJoinData(tc.j, tc.data)
			assert.Equal(t, tc.want, got)
		})
	}
}
