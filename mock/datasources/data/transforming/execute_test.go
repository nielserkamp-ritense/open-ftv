package transforming

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

func TestRunner_AddResult(t *testing.T) {
	t.Parallel()

	tr1 := &schema.Transformation{Object: schema.Object{Parent: schema.Parent{ID: "tr1"}}}
	tr2 := &schema.Transformation{Object: schema.Object{Parent: schema.Parent{ID: "tr2"}}}
	tr3 := &schema.Transformation{Object: schema.Object{Parent: schema.Parent{ID: "tr3"}}}

	t1 := &schema.Table{
		Object:     schema.Object{Parent: schema.Parent{ID: "t1"}},
		Transforms: []*schema.Transformation{tr1, tr2, tr3},
	}

	ds := &schema.Datasource{
		Parent: schema.Parent{ID: "ds"},
		Tables: []*schema.Table{t1},
	}
	ds.Fix(nil)

	testCases := []struct {
		name string
		r    *runner
		in   any
		id   string
	}{
		{
			name: "nil",
			r:    &runner{rec: map[string]any{}, transform: tr1},
			in:   nil,
			id:   "tr1",
		},
		{
			name: "string",
			r:    &runner{rec: map[string]any{}, transform: tr2},
			in:   "hello world",
			id:   "tr2",
		},
		{
			name: "qualified integer",
			r:    &runner{qualified: true, rec: map[string]any{}, transform: tr3},
			in:   123456,
			id:   "t1.tr3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.r.addResult(tc.in)
			got := tc.r.rec[tc.id]
			assert.Equal(t, tc.in, got)
		})
	}
}

func TestExecute(t *testing.T) {
	t.Parallel()

	fixedToday()

	tr1 := &schema.Transformation{
		Object:             schema.Object{Parent: schema.Parent{ID: "tr1"}},
		TransformationType: enums.TransformCompare,
		CompareType:        enums.IsNotEqual,
		ResultType:         enums.StringType,
		InputValues:        map[int]any{1: 12, 2: 13},
	}

	tr2 := &schema.Transformation{
		Object:             schema.Object{Parent: schema.Parent{ID: "tr2"}},
		TransformationType: enums.TransformConvert,
		ResultType:         enums.IntegerType,
		InputValues:        map[int]any{1: 1234.5678},
	}

	tr3 := &schema.Transformation{
		Object:             schema.Object{Parent: schema.Parent{ID: "tr3"}},
		TransformationType: enums.TransformAge,
		ResultType:         enums.IntegerType,
		InputValues:        map[int]any{1: 20011031},
	}

	t1 := &schema.Table{
		Object:     schema.Object{Parent: schema.Parent{ID: "t1"}},
		Transforms: []*schema.Transformation{tr1, tr2, tr3},
	}

	ds := &schema.Datasource{
		Parent: schema.Parent{ID: "ds"},
		Tables: []*schema.Table{t1},
	}
	ds.Fix(nil)

	testCases := []struct {
		name      string
		rec       map[string]any
		qualified bool
		t         *schema.Transformation
		want      map[string]any
	}{
		{
			name: "bad type",
			rec:  map[string]any{},
			t:    &schema.Transformation{},
			want: map[string]any{},
		},
		{
			name: "compare",
			rec:  map[string]any{},
			t:    tr1,
			want: map[string]any{"tr1": "true"},
		},
		{
			name: "convert",
			rec:  map[string]any{"tr2": 12},
			t:    tr2,
			want: map[string]any{"tr2": int64(1235)},
		},
		{
			name:      "age",
			rec:       map[string]any{"tr2": 12},
			qualified: true,
			t:         tr3,
			want:      map[string]any{"tr2": 12, "t1.tr3": int64(23)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Execute(tc.rec, tc.qualified, tc.t, nil)
			assert.EqualValues(t, tc.want, got)
		})
	}
}
