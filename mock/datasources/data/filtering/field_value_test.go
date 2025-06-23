package filtering

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

func TestExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		field  string
		exists bool
		table  string
		join   string
		want   *FieldValueFilter
	}{
		{
			name:   "exists - primary",
			field:  "f1",
			exists: true,
			want:   &FieldValueFilter{Level: enums.PrimaryLevel, Field: "f1", Compare: enums.Exists},
		},
		{
			name:   "not exists - primary",
			field:  "f2",
			exists: false,
			want:   &FieldValueFilter{Level: enums.PrimaryLevel, Field: "f2", Compare: enums.NotExists},
		},
		{
			name:   "exists - table",
			field:  "f3",
			exists: true,
			table:  "ouder2",
			want:   &FieldValueFilter{Level: enums.AnyLevel, Table: "ouder2", Field: "f3", Compare: enums.Exists},
		},
		{
			name:   "not exists - table",
			field:  "f4",
			exists: false,
			table:  "ouder1",
			want:   &FieldValueFilter{Level: enums.AnyLevel, Table: "ouder1", Field: "f4", Compare: enums.NotExists},
		},
		{
			name:   "exists - join",
			field:  "f5",
			exists: true,
			join:   "land",
			want:   &FieldValueFilter{Level: enums.JoinLevel, Join: "land", Field: "f5", Compare: enums.Exists},
		},
		{
			name:   "not exists - join",
			field:  "f6",
			exists: false,
			join:   "adres",
			want:   &FieldValueFilter{Level: enums.JoinLevel, Join: "adres", Field: "f6", Compare: enums.NotExists},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Exists(tc.field, tc.exists)
			require.NotNil(t, got)

			if tc.table != "" {
				got2 := got.OnAnyTable(tc.table)
				require.Equal(t, got, got2)
			}
			if tc.join != "" {
				got2 := got.OnJoin(tc.join)
				require.Equal(t, got, got2)
			}

			assert.Equal(t, tc.want.Level, got.Level)
			assert.False(t, got.Insensitive)
			assert.Equal(t, tc.want.Join, got.Join)
			assert.Equal(t, tc.want.Table, got.Table)
			assert.Equal(t, tc.want.Field, got.Field)
			assert.Equal(t, tc.want.Compare, got.Compare)
			assert.Nil(t, got.Value)
			assert.Nil(t, got.Values)
		})
	}
}

func TestCompare(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		field       string
		compare     enums.CompareType
		value       any
		insensitive bool
		want        *FieldValueFilter
	}{
		{
			name:    "exists - primary",
			field:   "f1",
			compare: enums.Exists,
			value:   1,
			want:    &FieldValueFilter{Level: enums.PrimaryLevel, Field: "f1", Compare: enums.Exists},
		},
		{
			name:    "not exists - primary",
			field:   "f2",
			compare: enums.NotExists,
			value:   "yo",
			want:    &FieldValueFilter{Level: enums.PrimaryLevel, Field: "f2", Compare: enums.NotExists},
		},
		{
			name:    "in list - primary",
			field:   "f3",
			compare: enums.InList,
			value:   "hello",
			want:    &FieldValueFilter{Level: enums.PrimaryLevel, Field: "f3", Compare: enums.IsEqual, Value: "hello"},
		},
		{
			name:    "not in list - primary",
			field:   "f4",
			compare: enums.NotInList,
			value:   "yo",
			want:    &FieldValueFilter{Level: enums.PrimaryLevel, Field: "f4", Compare: enums.IsNotEqual, Value: "yo"},
		},
		{
			name:        "is equal - insensitive - primary",
			field:       "f5",
			compare:     enums.IsEqual,
			value:       123,
			insensitive: true,
			want:        &FieldValueFilter{Level: enums.PrimaryLevel, Field: "f5", Compare: enums.IsEqual, Value: 123, Insensitive: true},
		},
		{
			name:    "like - sensitive - primary",
			field:   "f6",
			compare: enums.IsLike,
			value:   "%hello%",
			want:    &FieldValueFilter{Level: enums.PrimaryLevel, Field: "f6", Compare: enums.IsLike, Value: "%hello%"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Compare(tc.field, tc.compare, tc.value)
			require.NotNil(t, got)

			if tc.insensitive {
				got2 := got.CaseInsensitive()
				require.Equal(t, got, got2)
			} else {
				got2 := got.CaseSensitive()
				require.Equal(t, got, got2)
			}

			assert.Equal(t, tc.want.Level, got.Level)
			assert.Equal(t, tc.want.Insensitive, got.Insensitive)
			assert.Empty(t, got.Join)
			assert.Empty(t, got.Table)
			assert.Equal(t, tc.want.Field, got.Field)
			assert.Equal(t, tc.want.Compare, got.Compare)
			assert.Equal(t, tc.want.Value, got.Value)
			assert.Nil(t, got.Values)
		})
	}
}

func TestList(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		field  string
		exists bool
		values []any
		want   *FieldValueFilter
	}{
		{
			name:   "in list - primary",
			field:  "f1",
			exists: true,
			values: []any{1, 2, 3, 4},
			want:   &FieldValueFilter{Level: enums.PrimaryLevel, Field: "f1", Compare: enums.InList, Values: []any{1, 2, 3, 4}},
		},
		{
			name:   "not in list - primary",
			field:  "f2",
			values: []any{"yo", "ya", "yi", "yu"},
			want:   &FieldValueFilter{Level: enums.PrimaryLevel, Field: "f2", Compare: enums.NotInList, Values: []any{"yo", "ya", "yi", "yu"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := List(tc.field, tc.exists, tc.values...)
			require.NotNil(t, got)

			assert.Equal(t, tc.want.Level, got.Level)
			assert.False(t, got.Insensitive)
			assert.Empty(t, got.Join)
			assert.Empty(t, got.Table)
			assert.Equal(t, tc.want.Field, got.Field)
			assert.Equal(t, tc.want.Compare, got.Compare)
			assert.Nil(t, got.Value)
			assert.EqualValues(t, tc.want.Values, got.Values)
		})
	}
}

func TestFieldValueFilter_Prepare(t *testing.T) {
	t.Parallel()

	f1 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: enums.StringType}
	f2 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: enums.IntegerType}
	f3 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: enums.BooleanType}
	f4 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f4"}}, Type: enums.StringType}
	f5 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "geboortedatum"}}, Type: enums.DateType}

	tr1 := &schema.Transformation{
		Object:             schema.Object{Parent: schema.Parent{ID: "leeftijd"}},
		TransformationType: enums.TransformAge,
		ResultType:         enums.IntegerType,
		IsPII:              true,
		InputFields:        map[int]string{1: "geboortedatum"},
	}

	tr2 := &schema.Transformation{
		Object:             schema.Object{Parent: schema.Parent{ID: "volwassen"}},
		TransformationType: enums.TransformCompare,
		ResultType:         enums.BooleanType,
		CompareType:        enums.IsGreaterOrEqual,
		IsPII:              true,
		InputFields:        map[int]string{1: "leeftijd"},
		InputValues:        map[int]any{2: 18},
	}

	t1 := &schema.Table{
		Object: schema.Object{
			Parent: schema.Parent{ID: "t1"},
			Fields: []*schema.Field{f1, f2, f3, f5},
		},
		Transforms: []*schema.Transformation{tr1},
	}

	t2 := &schema.Table{
		Object: schema.Object{
			Parent: schema.Parent{ID: "t2"},
			Fields: []*schema.Field{f4, f1, f5},
		},
		Transforms: []*schema.Transformation{tr1, tr2},
	}

	ds := &schema.Datasource{
		Parent: schema.Parent{ID: "ds"},
		Tables: []*schema.Table{t1, t2},
	}

	ds.Fix(nil)

	j1 := &schema.Join{Target: "t1", Source: "t2", Fields: []string{"f1"}, JoinID: "j1"}
	j2 := &schema.Join{Target: "t1", Source: "xyz", Fields: []string{"f1"}, JoinID: "j2"}
	j3 := &schema.Join{Target: "t1", Source: "t2", Fields: []string{"xyz"}, JoinID: "j3"}

	j1.Fix(ds)
	j2.Fix(ds)
	j3.Fix(ds)

	testCases := []struct {
		name          string
		f             *FieldValueFilter
		ds            *schema.Datasource
		joins         []*schema.Join
		wantErr       bool
		wantTable     *schema.Table
		wantJoin      *schema.Join
		wantField     *schema.Field
		wantTransform *schema.Transformation
		wantRX        string
		wantList      map[string]struct{}
	}{
		{
			name:    "table and join",
			f:       &FieldValueFilter{Table: "table", Join: "join"},
			wantErr: true,
		},
		{
			name:    "value and values",
			f:       &FieldValueFilter{Value: "x", Values: []any{1, 2, 3}},
			wantErr: true,
		},
		{
			name:    "exists with a value",
			f:       &FieldValueFilter{Compare: enums.Exists, Value: "x"},
			wantErr: true,
		},
		{
			name:    "not exists with values",
			f:       &FieldValueFilter{Compare: enums.NotExists, Values: []any{1, 2, 3}},
			wantErr: true,
		},
		{
			name:    "in list without values",
			f:       &FieldValueFilter{Compare: enums.InList},
			wantErr: true,
		},
		{
			name:    "is equal without a value",
			f:       &FieldValueFilter{Compare: enums.IsEqual},
			wantErr: true,
		},
		{
			name:    "greater without a value",
			f:       &FieldValueFilter{Compare: enums.IsGreater},
			wantErr: true,
		},
		{
			name:    "like without a value",
			f:       &FieldValueFilter{Compare: enums.IsLike},
			wantErr: true,
		},
		{
			name:    "regex without a value",
			f:       &FieldValueFilter{Compare: enums.MatchRegex},
			wantErr: true,
		},
		{
			name:    "invalid compare",
			f:       &FieldValueFilter{Level: enums.PrimaryLevel, Compare: 250},
			wantErr: true,
		},
		{
			name:    "table at wrong level",
			f:       &FieldValueFilter{Level: enums.PrimaryLevel, Table: "table", Compare: enums.Exists},
			ds:      ds,
			wantErr: true,
		},
		{
			name:    "no table at any level",
			f:       &FieldValueFilter{Level: enums.AnyLevel, Compare: enums.Exists},
			ds:      ds,
			wantErr: true,
		},
		{
			name:    "unknown table",
			f:       &FieldValueFilter{Level: enums.AnyLevel, Table: "xyz", Compare: enums.Exists},
			ds:      ds,
			wantErr: true,
		},
		{
			name:    "good table, unknown field",
			f:       &FieldValueFilter{Level: enums.AnyLevel, Table: "t1", Field: "xyz", Compare: enums.Exists},
			ds:      ds,
			wantErr: true,
		},
		{
			name:      "good table, good field",
			f:         &FieldValueFilter{Level: enums.AnyLevel, Table: "t1", Field: "f1", Compare: enums.Exists},
			ds:        ds,
			wantTable: t1,
			wantField: f1,
		},
		{
			name:          "good table, good transform",
			f:             &FieldValueFilter{Level: enums.AnyLevel, Table: "t2", Field: "volwassen", Compare: enums.IsEqual, Value: true},
			ds:            ds,
			wantTable:     t2,
			wantTransform: tr2,
		},
		{
			name:    "join at wrong level",
			f:       &FieldValueFilter{Level: enums.PrimaryLevel, Join: "join", Compare: enums.Exists},
			ds:      ds,
			wantErr: true,
		},
		{
			name:    "no join at join level",
			f:       &FieldValueFilter{Level: enums.JoinLevel, Compare: enums.Exists},
			ds:      ds,
			wantErr: true,
		},
		{
			name:    "unknown join",
			f:       &FieldValueFilter{Level: enums.JoinLevel, Join: "xyz", Compare: enums.Exists},
			ds:      ds,
			joins:   []*schema.Join{j1, j2, j3},
			wantErr: true,
		},
		{
			name:    "good join, bad source",
			f:       &FieldValueFilter{Level: enums.JoinLevel, Join: "j2", Compare: enums.Exists},
			ds:      ds,
			joins:   []*schema.Join{j1, j2, j3},
			wantErr: true,
		},
		{
			name:    "good join, bad field",
			f:       &FieldValueFilter{Level: enums.JoinLevel, Join: "J3", Compare: enums.Exists},
			ds:      ds,
			joins:   []*schema.Join{j1, j2, j3},
			wantErr: true,
		},
		{
			name:      "good join",
			f:         &FieldValueFilter{Level: enums.JoinLevel, Join: "J1", Field: "f1", Compare: enums.Exists},
			ds:        ds,
			joins:     []*schema.Join{j1, j2, j3},
			wantJoin:  j1,
			wantTable: t2,
			wantField: f1,
		},
		{
			name:    "duplicate field",
			f:       &FieldValueFilter{Field: "f1", Compare: enums.Exists},
			ds:      ds,
			wantErr: true,
		},
		{
			name:    "duplicate transform",
			f:       &FieldValueFilter{Field: "leeftijd", Compare: enums.Exists},
			ds:      ds,
			wantErr: true,
		},
		{
			name:    "unknown field table",
			f:       &FieldValueFilter{Field: "xyz.f1", Compare: enums.Exists},
			ds:      ds,
			wantErr: true,
		},
		{
			name:    "unknown field",
			f:       &FieldValueFilter{Field: "xyz", Compare: enums.Exists},
			ds:      ds,
			wantErr: true,
		},
		{
			name:      "in list with values",
			f:         &FieldValueFilter{Field: "f2", Compare: enums.InList, Values: []any{1, 2.34, true}},
			ds:        ds,
			wantTable: t1,
			wantField: f2,
			wantList:  map[string]struct{}{"1": {}, "2.34": {}, "true": {}},
		},
		{
			name:          "equal with value",
			f:             &FieldValueFilter{Field: "volwassen", Compare: enums.IsEqual, Value: true},
			ds:            ds,
			wantTable:     t2,
			wantTransform: tr2,
		},
		{
			name:          "exists",
			f:             &FieldValueFilter{Field: "t2.leeftijd", Compare: enums.Exists},
			ds:            ds,
			wantTable:     t2,
			wantTransform: tr1,
		},
		{
			name:      "like with a value",
			f:         &FieldValueFilter{Field: "t1.f1", Compare: enums.IsLike, Value: "%hello%"},
			ds:        ds,
			wantTable: t1,
			wantField: f1,
			wantRX:    `^.*hello.*$`,
		},
		{
			name:      "not regex with a value",
			f:         &FieldValueFilter{Field: "t2.f4", Compare: enums.NotMatchRegex, Value: ".*lo.?wo.*"},
			ds:        ds,
			wantTable: t2,
			wantField: f4,
			wantRX:    `^.*lo.?wo.*$`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.f.Prepare(tc.ds, tc.joins)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantTable, tc.f.table)
				assert.Equal(t, tc.wantJoin, tc.f.join)
				assert.Equal(t, tc.wantField, tc.f.field)
				assert.Equal(t, tc.wantTransform, tc.f.transform)

				if tc.wantRX != "" {
					assert.Equal(t, tc.wantRX, tc.f.rx.String())
				}

				assert.EqualValues(t, tc.wantList, tc.f.list)
			}
		})
	}
}

func TestFieldValueFilter_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		f    *FieldValueFilter
		want string
	}{
		{
			name: "primary equal",
			f: &FieldValueFilter{
				Level:       enums.PrimaryLevel,
				Insensitive: true,
				Field:       "voornaam",
				Compare:     enums.IsEqual,
				Value:       "pieter",
			},
			want: "{level=primary,case-insensitive,field=voornaam,compare-type=IsEqual,value=pieter}",
		},
		{
			name: "primary in list",
			f: &FieldValueFilter{
				Level:       enums.PrimaryLevel,
				Insensitive: true,
				Field:       "voornaam",
				Compare:     enums.InList,
				Values:      []any{"pieter", "piet", "pietje"},
			},
			want: "{level=primary,case-insensitive,field=voornaam,compare-type=InList,values=[pieter piet pietje]}",
		},
		{
			name: "primary exists",
			f: &FieldValueFilter{
				Level:       enums.PrimaryLevel,
				Insensitive: true,
				Field:       "bsn",
				Compare:     enums.Exists,
			},
			want: "{level=primary,case-insensitive,field=bsn,compare-type=Exists}",
		},
		{
			name: "join not equal",
			f: &FieldValueFilter{
				Level:       enums.JoinLevel,
				Insensitive: true,
				Join:        "adres",
				Field:       "postcode",
				Compare:     enums.IsNotEqual,
				Value:       "9999zz",
			},
			want: "{level=join,case-insensitive,join=adres,field=postcode,compare-type=IsNotEqual,value=9999zz}",
		},
		{
			name: "table regex",
			f: &FieldValueFilter{
				Level:   enums.AnyLevel,
				Table:   "adres",
				Field:   "postcode",
				Compare: enums.MatchRegex,
				Value:   "(1111|2222|3333)[a-z][a-z]",
			},
			want: "{level=any,table=adres,field=postcode,compare-type=MatchRegex,value=(1111|2222|3333)[a-z][a-z]}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.f.String()
			assert.Equal(t, tc.want, got)
		})
	}
}
