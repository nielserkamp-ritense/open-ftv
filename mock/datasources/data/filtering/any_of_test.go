package filtering

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

func TestAnyOfFilter_MatchOnPrimaryData(t *testing.T) {
	t.Parallel()

	f1 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: enums.StringType}
	f2 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: enums.IntegerType}
	f3 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: enums.BooleanType}
	f4 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f4"}}, Type: enums.StringType}

	t1 := &schema.Table{
		Object: schema.Object{
			Parent: schema.Parent{ID: "t1"},
			Fields: []*schema.Field{f1, f2, f3},
		},
	}

	t2 := &schema.Table{
		Object: schema.Object{
			Parent: schema.Parent{ID: "t2"},
			Fields: []*schema.Field{f4, f1},
		},
	}

	ds := &schema.Datasource{
		Parent: schema.Parent{ID: "ds"},
		Tables: []*schema.Table{t1, t2},
	}

	ds.Fix(nil)

	j1 := &schema.Join{Source: "t1", JoinID: "j1"}
	j1.Fix(ds)

	j2 := &schema.Join{Source: "t2", JoinID: "j2"}
	j2.Fix(ds)

	testCases := []struct {
		name    string
		f       AnyOfFilter
		ds      *schema.Datasource
		joins   []*schema.Join
		data    map[string]any
		wantErr bool
		want    bool
	}{
		{
			name: "prepare fails (table missing)",
			f: AnyOfFilter{AnyOf: Filters{
				&Filter{FieldValueFilter: FieldValueFilter{Level: enums.AnyLevel, Field: "f1"}}},
			},
			ds:      ds,
			wantErr: true,
		},
		{
			name: "prepare fails (join missing)",
			f: AnyOfFilter{AnyOf: Filters{
				&Filter{FieldValueFilter: FieldValueFilter{Level: enums.JoinLevel, Join: "xyz", Field: "f1"}}},
			},
			ds:      ds,
			joins:   []*schema.Join{j1, j2},
			wantErr: true,
		},
		{
			name: "filters match",
			f: AnyOfFilter{AnyOf: Filters{
				&Filter{FieldValueFilter: FieldValueFilter{Level: enums.PrimaryLevel, Field: "t1.f1", Compare: enums.IsLesser, Value: "1"}},
				&Filter{FieldValueFilter: FieldValueFilter{Level: enums.PrimaryLevel, Field: "t1.f1", Compare: enums.Exists}},
			}},
			ds:   ds,
			data: map[string]any{"f1": 2},
			want: true,
		},
		{
			name: "filters do not match",
			f: AnyOfFilter{AnyOf: Filters{
				&Filter{FieldValueFilter: FieldValueFilter{Level: enums.PrimaryLevel, Field: "t1.f1", Compare: enums.IsEqual, Value: "2"}},
				&Filter{FieldValueFilter: FieldValueFilter{Level: enums.PrimaryLevel, Field: "f2", Compare: enums.NotExists}},
				&Filter{FieldValueFilter: FieldValueFilter{Level: enums.PrimaryLevel, Field: "f2", Compare: enums.IsGreater, Value: "1"}},
			}},
			ds:   ds,
			data: map[string]any{"f1": 1, "f2": "hello"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := tc.f
			err := f.Prepare(tc.ds, tc.joins)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				got := f.MatchOnPrimaryData(tc.data)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestAnyOfFilter_MatchOnJoinData(t *testing.T) {
	t.Parallel()

	f1 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: enums.StringType}
	f2 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: enums.IntegerType}
	f3 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: enums.BooleanType}
	f4 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f4"}}, Type: enums.StringType}

	t1 := &schema.Table{
		Object: schema.Object{
			Parent: schema.Parent{ID: "t1"},
			Fields: []*schema.Field{f1, f2, f3},
		},
	}

	t2 := &schema.Table{
		Object: schema.Object{
			Parent: schema.Parent{ID: "t2"},
			Fields: []*schema.Field{f4, f1},
		},
	}

	ds := &schema.Datasource{
		Parent: schema.Parent{ID: "ds"},
		Tables: []*schema.Table{t1, t2},
	}

	ds.Fix(nil)

	j1 := &schema.Join{Source: "t1", JoinID: "j1"}
	j1.Fix(ds)

	j2 := &schema.Join{Source: "t2", JoinID: "j2"}
	j2.Fix(ds)

	testCases := []struct {
		name    string
		f       AnyOfFilter
		ds      *schema.Datasource
		joins   []*schema.Join
		j       *schema.Join
		data    map[string]any
		wantErr bool
		want    bool
	}{
		{
			name: "prepare fails (table missing)",
			f: AnyOfFilter{AnyOf: Filters{
				&Filter{FieldValueFilter: FieldValueFilter{Level: enums.AnyLevel, Field: "f1"}}},
			},
			ds:      ds,
			wantErr: true,
		},
		{
			name: "prepare fails (join missing)",
			f: AnyOfFilter{AnyOf: Filters{
				&Filter{FieldValueFilter: FieldValueFilter{Level: enums.JoinLevel, Join: "xyz", Field: "f1"}}},
			},
			ds:      ds,
			joins:   []*schema.Join{j1, j2},
			wantErr: true,
		},
		{
			name: "filters match",
			f: AnyOfFilter{AnyOf: Filters{
				&Filter{FieldValueFilter: FieldValueFilter{Level: enums.JoinLevel, Join: "j1", Field: "f1", Compare: enums.IsLesser, Value: "2"}},
				&Filter{FieldValueFilter: FieldValueFilter{Level: enums.JoinLevel, Join: "j1", Field: "f1", Compare: enums.Exists}},
			}},
			ds:    ds,
			joins: []*schema.Join{j1, j2},
			j:     j1,
			data:  map[string]any{"f1": 2},
			want:  true,
		},
		{
			name: "filters do not match",
			f: AnyOfFilter{AnyOf: Filters{
				&Filter{FieldValueFilter: FieldValueFilter{Level: enums.AnyLevel, Table: "t1", Field: "f1", Compare: enums.IsEqual, Value: "2"}},
				&Filter{FieldValueFilter: FieldValueFilter{Level: enums.JoinLevel, Join: "j1", Field: "f2", Compare: enums.NotExists}},
				&Filter{FieldValueFilter: FieldValueFilter{Level: enums.JoinLevel, Join: "j1", Field: "f2", Compare: enums.IsGreater, Value: "1"}},
			}},
			ds:    ds,
			joins: []*schema.Join{j1, j2},
			j:     j1,
			data:  map[string]any{"f1": 1, "f2": "hello"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			f := tc.f
			err := f.Prepare(tc.ds, tc.joins)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				got := f.MatchOnJoinData(tc.j, tc.data)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}
