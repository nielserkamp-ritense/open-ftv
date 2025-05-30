package models

import (
	"regexp"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/types"
)

func TestRow_FieldString(t *testing.T) {
	t.Parallel()

	r1 := &Row{
		Data: map[string]any{
			"foo":       "bar",
			"int":       123,
			"float":     567.89,
			"bool":      true,
			"timestamp": time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	testCases := []struct {
		name string
		row  *Row
		id   string
		want string
	}{
		{name: "empty", row: r1},
		{name: "not found", row: r1, id: "yo"},
		{name: "record empty", row: &Row{Data: map[string]any{}}, id: "foo"},
		{name: "string", row: r1, id: "foo", want: "bar"},
		{name: "int", row: r1, id: "int", want: "123"},
		{name: "float", row: r1, id: "float", want: "567.89"},
		{name: "bool", row: r1, id: "bool", want: "true"},
		{name: "timestamp", row: r1, id: "timestamp", want: "2020-02-01 00:00:00 +0000 UTC"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.row.FieldString(tc.id)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestRowFromData(t *testing.T) {
	t.Parallel()

	t1 := &schema.Object{
		Parent:      schema.Parent{ID: "t1"},
		Description: "t1",
		Fields: []*schema.Field{
			{
				Object: schema.Object{Parent: schema.Parent{ID: "f1"}},
				Type:   types.StringType,
				IsPII:  true,
			},
			{
				Object: schema.Object{Parent: schema.Parent{ID: "f2"}},
				Type:   types.IntegerType,
			},
			{
				Object: schema.Object{Parent: schema.Parent{ID: "f3"}},
				Type:   types.DateType,
				IsPII:  true,
			},
			{
				Object: schema.Object{Parent: schema.Parent{ID: "f4"}},
				Type:   types.EmailType,
				IsPII:  true,
			},
		},
	}

	testCases := []struct {
		name string
		t    *schema.Object
		data map[string]any
		want *Row
	}{
		{
			name: "all empty",
			t:    t1,
			data: map[string]any{},
			want: &Row{Data: map[string]any{}},
		},
		{
			name: "unknown fields",
			t:    t1,
			data: map[string]any{"hello": "world"},
			want: &Row{Data: map[string]any{}},
		},
		{
			name: "f1",
			t:    t1,
			data: map[string]any{"f1": "hello world"},
			want: &Row{Data: map[string]any{"f1": "hello world"}},
		},
		{
			name: "all",
			t:    t1,
			data: map[string]any{
				"f1": "hello world",
				"f2": 123,
				"f3": time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				"f4": "joep@tv.nl",
			},
			want: &Row{
				Data: map[string]any{
					"f1": "hello world",
					"f2": int64(123),
					"f3": time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
					"f4": "joep@tv.nl",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := RowFromData(tc.t, tc.data)
			assert.EqualValues(t, tc.want.Data, got.Data)
		})
	}
}

func TestRowFromCSV(t *testing.T) {
	t.Parallel()

	t1 := &schema.Object{
		Parent:      schema.Parent{ID: "t1"},
		Description: "t1",
		Fields: []*schema.Field{
			{
				Object: schema.Object{Parent: schema.Parent{ID: "f1"}},
				Type:   types.StringType,
				IsPII:  true,
			},
			{
				Object: schema.Object{Parent: schema.Parent{ID: "f2"}},
				Type:   types.IntegerType,
			},
			{
				Object: schema.Object{Parent: schema.Parent{ID: "f3"}},
				Type:   types.DateType,
				IsPII:  true,
			},
			{
				Object: schema.Object{Parent: schema.Parent{ID: "f4"}},
				Type:   types.EmailType,
				IsPII:  true,
			},
		},
	}

	testCases := []struct {
		name    string
		t       *schema.Object
		headers []string
		data    []string
		want    *Row
	}{
		{
			name:    "all empty",
			t:       t1,
			headers: []string{},
			data:    []string{},
			want:    &Row{Data: map[string]any{}},
		},
		{
			name:    "unknown fields",
			t:       t1,
			headers: []string{"hello"},
			data:    []string{"world"},
			want:    &Row{Data: map[string]any{}},
		},
		{
			name:    "f1",
			t:       t1,
			headers: []string{"f1"},
			data:    []string{"hello world"},
			want:    &Row{Data: map[string]any{"f1": "hello world"}},
		},
		{
			name:    "all",
			t:       t1,
			headers: []string{"f1", "f2", "f3", "f4"},
			data:    []string{"hello world", "123", "2024-06-10", "joep@tv.nl"},
			want: &Row{
				Data: map[string]any{
					"f1": "hello world",
					"f2": int64(123),
					"f3": time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC),
					"f4": "joep@tv.nl",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := RowFromCSV(tc.t, tc.headers, tc.data)
			assert.EqualValues(t, tc.want.Data, got.Data)
		})
	}
}

func TestRow_MatchFilter(t *testing.T) {
	t.Parallel()

	var r1 = &Row{Data: map[string]any{
		"f1": "hello world",
		"f2": 123,
		"f3": time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		"f4": true,
		"f5": 98765.4321,
	}}

	testCases := []struct {
		name   string
		row    *Row
		filter map[string]any
		want   bool
	}{
		{
			name: "empty filter",
			row:  r1,
			want: true,
		},
		{
			name:   "matched string",
			row:    r1,
			filter: map[string]any{"f1": "hello world"},
			want:   true,
		},
		{
			name:   "unmatched string",
			row:    r1,
			filter: map[string]any{"f1": "hello worlds"},
		},
		{
			name:   "matched regex",
			row:    r1,
			filter: map[string]any{"f1": regexp.MustCompile(".*ello.*")},
			want:   true,
		},
		{
			name:   "unmatched regex",
			row:    r1,
			filter: map[string]any{"f1": regexp.MustCompile(".*elo.*")},
		},
		{
			name:   "matched int",
			row:    r1,
			filter: map[string]any{"f2": "123"},
			want:   true,
		},
		{
			name:   "unmatched int",
			row:    r1,
			filter: map[string]any{"f2": 124},
		},
		{
			name:   "matched date",
			row:    r1,
			filter: map[string]any{"f3": time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)},
			want:   true,
		},
		{
			name:   "unmatched date",
			row:    r1,
			filter: map[string]any{"f3": time.Date(2024, 6, 1, 0, 0, 0, 1, time.UTC)},
		},
		{
			name:   "matched float",
			row:    r1,
			filter: map[string]any{"f5": 98765.4321},
			want:   true,
		},
		{
			name:   "unmatched float",
			row:    r1,
			filter: map[string]any{"f5": 98765.432},
		},
		{
			name: "all matched",
			row:  r1,
			filter: map[string]any{
				"f5": 98765.4321,
				"f2": "123",
				"f1": regexp.MustCompile(".*ello.*"),
				"f4": true,
				"f3": regexp.MustCompile("2024-06*"),
			},
			want: true,
		},
		{
			name: "one unmatched",
			row:  r1,
			filter: map[string]any{
				"f3": time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				"f2": "123",
				"f1": regexp.MustCompile(".*ello.*"),
				"f4": "1", // this is not "true"
				"f5": 98765.4321},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.row.MatchFilter(tc.filter)
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestRow_MatchFields(t *testing.T) {
	t.Parallel()

	def := makeRowDatasource1()

	r1 := &Row{Data: map[string]any{"f1": 1, "f2": "hello world", "f3": true, "f4": 1.2345}, def: def}
	r2 := &Row{Data: map[string]any{"f1": 2, "f2": "hello mars", "f4": 12.345}, def: def}
	r3 := &Row{Data: map[string]any{"f1": 3, "f2": "goodbye world", "f3": true, "f4": 123.45}, def: def}

	testCases := []struct {
		name    string
		row     *Row
		matcher types.FieldMatcher
		want    *Row
	}{
		{
			name:    "no match",
			row:     r1,
			matcher: types.NewFieldMatcher("f5,f6,f7"),
			want:    &Row{Data: make(map[string]any), def: def},
		},
		{
			name:    "match single field",
			row:     r2,
			matcher: types.NewFieldMatcher("f5,f6,f1"),
			want:    &Row{Data: map[string]any{"f1": 2}, def: def},
		},
		{
			name:    "match few fields",
			row:     r3,
			matcher: types.NewFieldMatcher("f5,f2,ab*,f4"),
			want:    &Row{Data: map[string]any{"f2": "goodbye world", "f4": 123.45}, def: def},
		},
		{
			name:    "match all fields",
			row:     r3,
			matcher: types.NewFieldMatcher("f5,x?,,*"),
			want:    &Row{Data: r3.Data, def: def},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.row.MatchFields(tc.matcher)
			assert.Equal(t, tc.want.def, got.def)
			assert.EqualValues(t, tc.want.Data, got.Data)
		})
	}
}

func TestRow_JoinSibling(t *testing.T) {
	t.Parallel()

	def1 := makeRowDatasource1()
	def2 := makeRowDatasource2()

	r1 := &Row{Data: map[string]any{"f1": 1, "f2": "hello world", "f3": true, "f4": 1.2345}, def: def1}
	r2 := &Row{Data: map[string]any{"f5": 2, "f6": "hello mars", "f8": 12.345}, def: def2}
	r3 := &Row{Data: map[string]any{"f5": 2, "f2": "hello mars", "f8": 12.345}, def: def2}

	testCases := []struct {
		name string
		r    *Row
		r2   *Row
		want *Row
	}{
		{
			name: "no overwrite",
			r:    r1,
			r2:   r2,
			want: &Row{Data: map[string]any{"f1": 1, "f2": "hello world", "f3": true, "f4": 1.2345, "f5": 2, "f6": "hello mars", "f8": 12.345}, def: def1},
		},
		{
			name: "overwrite",
			r:    r1,
			r2:   r3,
			want: &Row{Data: map[string]any{"f1": 1, "f2": "hello mars", "f3": true, "f4": 1.2345, "f5": 2, "f8": 12.345}, def: def1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.r.JoinSibling(tc.r2)
			assert.Equal(t, def1, got.def)
			assert.EqualValues(t, tc.want.Data, got.Data)
		})
	}
}

func TestRow_JoinSiblingQualified(t *testing.T) {
	t.Parallel()

	def1 := makeRowDatasource1()
	def2 := makeRowDatasource2()

	r1 := &Row{Data: map[string]any{"f1": 1, "f2": "hello world", "f3": true, "f4": 1.2345}, def: def1}
	r2 := &Row{Data: map[string]any{"f5": 2, "f6": "hello mars", "f8": 12.345}, def: def2}

	testCases := []struct {
		name string
		r    *Row
		r2   *Row
		want *Row
	}{
		{
			name: "both unqualified",
			r:    r1,
			r2:   r2,
			want: &Row{Data: map[string]any{"t1.f1": 1, "t1.f2": "hello world", "t1.f3": true, "t1.f4": 1.2345, "t2.f5": 2, "t2.f6": "hello mars", "t2.f8": 12.345}, def: def1},
		},
		{
			name: "primary qualified",
			r:    r1.Qualified("t1"),
			r2:   r2,
			want: &Row{Data: map[string]any{"t1.f1": 1, "t1.f2": "hello world", "t1.f3": true, "t1.f4": 1.2345, "t2.f5": 2, "t2.f6": "hello mars", "t2.f8": 12.345}, def: def1},
		},
		{
			name: "secondary qualified",
			r:    r1,
			r2:   r2.Qualified("t2"),
			want: &Row{Data: map[string]any{"t1.f1": 1, "t1.f2": "hello world", "t1.f3": true, "t1.f4": 1.2345, "t2.f5": 2, "t2.f6": "hello mars", "t2.f8": 12.345}, def: def1},
		},
		{
			name: "both qualified",
			r:    r1.Qualified("t1"),
			r2:   r2.Qualified("t2"),
			want: &Row{Data: map[string]any{"t1.f1": 1, "t1.f2": "hello world", "t1.f3": true, "t1.f4": 1.2345, "t2.f5": 2, "t2.f6": "hello mars", "t2.f8": 12.345}, def: def1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.r.JoinSiblingQualified("t1", "t2", tc.r2)
			assert.Equal(t, def1, got.def)
			assert.EqualValues(t, tc.want.Data, got.Data)
		})
	}
}

func TestRow_Qualified(t *testing.T) {
	t.Parallel()

	def1 := makeRowDatasource1()
	def2 := makeRowDatasource2()

	r1 := &Row{Data: map[string]any{"f1": 1, "f2": "hello world", "f3": true, "f4": 1.2345}, def: def1}
	r2 := &Row{Data: map[string]any{"f5": 2, "f6": "hello mars", "f8": 12.345}, def: def2}

	testCases := []struct {
		name string
		r    *Row
		t    string
		want *Row
	}{
		{
			name: "unqualified",
			r:    r1,
			t:    "t1",
			want: &Row{Data: map[string]any{"t1.f1": 1, "t1.f2": "hello world", "t1.f3": true, "t1.f4": 1.2345}, def: def1},
		},
		{
			name: "qualified",
			r:    r2.Qualified("t3"),
			t:    "t2",
			want: &Row{Data: map[string]any{"t3.f5": 2, "t3.f6": "hello mars", "t3.f8": 12.345}, def: def2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.r.Qualified(tc.t)
			assert.Equal(t, tc.want.def, got.def)
			assert.EqualValues(t, tc.want.Data, got.Data)
		})
	}
}

func TestRow_RemoveFieldMap(t *testing.T) {
	t.Parallel()

	def1 := makeRowDatasource1()
	r1 := &Row{Data: map[string]any{"f1": 1, "f2": "hello world", "f3": true, "f4": 1.2345}, def: def1}

	testCases := []struct {
		name   string
		r      *Row
		fields map[string]struct{}
		want   *Row
	}{
		{
			name:   "empty",
			r:      r1,
			fields: map[string]struct{}{},
			want:   &Row{Data: r1.Data, def: def1},
		},
		{
			name:   "single field, no match",
			r:      r1,
			fields: map[string]struct{}{"f5": {}},
			want:   &Row{Data: r1.Data, def: def1},
		},
		{
			name:   "single field, matched",
			r:      r1,
			fields: map[string]struct{}{"f4": {}},
			want:   &Row{Data: map[string]any{"f1": 1, "f2": "hello world", "f3": true}, def: def1},
		},
		{
			name:   "few fields, not matched",
			r:      r1,
			fields: map[string]struct{}{"f5": {}, "f6": {}},
			want:   &Row{Data: r1.Data, def: def1},
		},
		{
			name:   "few fields, some matched",
			r:      r1,
			fields: map[string]struct{}{"f2": {}, "f6": {}, "f4": {}},
			want:   &Row{Data: map[string]any{"f1": 1, "f3": true}, def: def1},
		},
		{
			name:   "few fields, all matched",
			r:      r1,
			fields: map[string]struct{}{"f4": {}, "f3": {}, "f2": {}},
			want:   &Row{Data: map[string]any{"f1": 1}, def: def1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.r.RemoveFieldMap(tc.fields)
			assert.Equal(t, tc.want.def, got.def)
			assert.EqualValues(t, tc.want.Data, got.Data)
		})
	}
}

func TestRow_RemoveFields(t *testing.T) {
	t.Parallel()

	def1 := makeRowDatasource1()
	r1 := &Row{Data: map[string]any{"f1": 1, "f2": "hello world", "f3": true, "f4": 1.2345}, def: def1}

	testCases := []struct {
		name   string
		r      *Row
		fields []string
		want   *Row
	}{
		{
			name:   "empty",
			r:      r1,
			fields: []string{},
			want:   &Row{Data: r1.Data, def: def1},
		},
		{
			name:   "single field, no match",
			r:      r1,
			fields: []string{"f5"},
			want:   &Row{Data: r1.Data, def: def1},
		},
		{
			name:   "single field, matched",
			r:      r1,
			fields: []string{"f4"},
			want:   &Row{Data: map[string]any{"f1": 1, "f2": "hello world", "f3": true}, def: def1},
		},
		{
			name:   "few fields, not matched",
			r:      r1,
			fields: []string{"f5", "f6"},
			want:   &Row{Data: r1.Data, def: def1},
		},
		{
			name:   "few fields, some matched",
			r:      r1,
			fields: []string{"f2", "f6", "f4"},
			want:   &Row{Data: map[string]any{"f1": 1, "f3": true}, def: def1},
		},
		{
			name:   "few fields, all matched",
			r:      r1,
			fields: []string{"f4", "f3", "f2"},
			want:   &Row{Data: map[string]any{"f1": 1}, def: def1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.r.RemoveFields(tc.fields)
			assert.Equal(t, tc.want.def, got.def)
			assert.EqualValues(t, tc.want.Data, got.Data)
		})
	}
}

func TestRow_JSON(t *testing.T) {
	t.Parallel()

	def1 := makeRowDatasource1()
	def2 := makeRowDatasource2()

	r1 := &Row{Data: map[string]any{"f1": float64(1), "f2": "hello world", "f3": true, "f4": 1.2345}, def: def1}
	r2 := &Row{Data: map[string]any{"f5": float64(2), "f6": "hello mars", "f8": 12.345}, def: def2}
	r3 := &Row{Data: map[string]any{"f5": float64(3), "f2": "hello mars", "f8": 12.345}, def: def2}

	testCases := []struct {
		name string
		r    *Row
	}{
		{name: "a", r: r1},
		{name: "b", r: r2},
		{name: "c", r: r3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(tc.r)
			require.NoError(t, err)
			require.NotEmpty(t, got)

			got2 := new(Row)
			err = json.Unmarshal(got, got2)
			require.NoError(t, err)

			assert.EqualValues(t, tc.r.Data, got2.Data)
		})
	}
}

func TestRow_UnmarshalJSON_Fail(t *testing.T) {
	t.Parallel()

	t.Run("unmarshal JSON fail", func(t *testing.T) {
		t.Parallel()

		r := new(Row)
		err := r.UnmarshalJSON([]byte("\000\001"))
		require.Error(t, err)
		require.Nil(t, r.Data)
	})
}

func TestRow_YAML(t *testing.T) {
	t.Parallel()

	def1 := makeRowDatasource1()
	def2 := makeRowDatasource2()

	r1 := &Row{Data: map[string]any{"f1": float64(1), "f2": "hello world", "f3": true, "f4": 1.2345}, def: def1}
	r2 := &Row{Data: map[string]any{"f5": float64(2), "f6": "hello mars", "f8": 12.345}, def: def2}
	r3 := &Row{Data: map[string]any{"f5": float64(3), "f2": "hello mars", "f8": 12.345}, def: def2}

	testCases := []struct {
		name string
		r    *Row
	}{
		{name: "a", r: r1},
		{name: "b", r: r2},
		{name: "c", r: r3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := yaml.Marshal(tc.r)
			require.NoError(t, err)
			require.NotEmpty(t, got)

			got2 := new(Row)
			err = yaml.Unmarshal(got, got2)
			require.NoError(t, err)

			assert.EqualValues(t, tc.r.Data, got2.Data)
		})
	}
}

func TestRow_UnmarshalYAML_Fail(t *testing.T) {
	t.Parallel()

	t.Run("unmarshal YAML fail", func(t *testing.T) {
		t.Parallel()

		r := new(Row)
		err := r.UnmarshalYAML([]byte("\000\001"))
		require.Error(t, err)
		require.Nil(t, r.Data)
	})
}

func TestRow_MarshalCSV(t *testing.T) {
	t.Parallel()

	t.Run("csv", func(t *testing.T) {
		t.Parallel()

		def1 := makeRowDatasource1()
		r1 := &Row{Data: map[string]any{"f1": float64(1), "f2": "hello world", "f3": true, "f4": 1.2345}, def: def1}

		got, err := r1.MarshalCSV()
		require.NoError(t, err)
		require.NotEmpty(t, got)

		want := `"f1","f2","f3","f4"
1,"hello world","true",1.2345
`

		assert.Equal(t, want, string(got))
	})
}

func TestRemoveNil(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		m    map[string]any
		want map[string]any
	}{
		{
			name: "none",
			m:    map[string]any{"f1": float64(1), "f2": "hello world", "f3": true, "f4": 1.2345},
			want: map[string]any{"f1": float64(1), "f2": "hello world", "f3": true, "f4": 1.2345},
		},
		{
			name: "single",
			m:    map[string]any{"f1": float64(1), "f2": nil, "f3": true, "f4": 1.2345},
			want: map[string]any{"f1": float64(1), "f3": true, "f4": 1.2345},
		},
		{
			name: "few",
			m:    map[string]any{"f1": float64(1), "f2": nil, "f3": nil, "f4": 1.2345, "f5": nil},
			want: map[string]any{"f1": float64(1), "f4": 1.2345},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := removeNil(tc.m)
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func makeRowDatasource1() *schema.Object {
	f1 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: types.IntegerType}
	f2 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: types.StringType}
	f3 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: types.BooleanType}
	f4 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f4"}}, Type: types.FloatType}

	t1 := &schema.Table{Object: schema.Object{
		Parent: schema.Parent{ID: "t1"},
		Fields: []*schema.Field{f1, f2, f3, f4},
	}}

	ds1 := &schema.Datasource{
		Parent: schema.Parent{ID: "ds1"},
		Tables: []*schema.Table{t1},
	}
	ds1.Fix(nil)

	return &t1.Object
}

func makeRowDatasource2() *schema.Object {
	f5 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f5"}}, Type: types.IntegerType}
	f6 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f6"}}, Type: types.StringType}
	f7 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f7"}}, Type: types.BooleanType}
	f8 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f8"}}, Type: types.FloatType}

	t2 := &schema.Table{Object: schema.Object{
		Parent: schema.Parent{ID: "t2"},
		Fields: []*schema.Field{f5, f6, f7, f8},
	}}

	ds2 := &schema.Datasource{
		Parent: schema.Parent{ID: "ds2"},
		Tables: []*schema.Table{t2},
	}
	ds2.Fix(nil)

	return &t2.Object
}
