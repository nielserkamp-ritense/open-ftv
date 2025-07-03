package transforming

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
)

func TestConvertList(t *testing.T) {
	t.Parallel()

	dt1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	dt2 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dt3 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name string
		in   any
		want []any
	}{
		{
			name: "int",
			in:   123,
			want: []any{123},
		},
		{
			name: "[]any",
			in:   []any{1, true, "hello world"},
			want: []any{1, true, "hello world"},
		},
		{
			name: "[]string",
			in:   []string{"hello", "world"},
			want: []any{"hello", "world"},
		},
		{
			name: "[]int",
			in:   []int{1, 9999999, 54321, -9},
			want: []any{1, 9999999, 54321, -9},
		},
		{
			name: "[]int8",
			in:   []int8{0, 1, -1, 9, -9},
			want: []any{int8(0), int8(1), int8(-1), int8(9), int8(-9)},
		},
		{
			name: "[]int16",
			in:   []int16{0, 1, -1, 9, -9},
			want: []any{int16(0), int16(1), int16(-1), int16(9), int16(-9)},
		},
		{
			name: "[]int32",
			in:   []int32{0, 1, -1, 9, -9},
			want: []any{int32(0), int32(1), int32(-1), int32(9), int32(-9)},
		},
		{
			name: "[]int64",
			in:   []int64{0, 1, -1, 9, -9},
			want: []any{int64(0), int64(1), int64(-1), int64(9), int64(-9)},
		},
		{
			name: "[]uint",
			in:   []uint{1, 9999999, 54321, 9},
			want: []any{uint(1), uint(9999999), uint(54321), uint(9)},
		},
		{
			name: "[]uint8",
			in:   []uint8{0, 1, 2, 9, 55},
			want: []any{uint8(0), uint8(1), uint8(2), uint8(9), uint8(55)},
		},
		{
			name: "[]uint16",
			in:   []uint16{0, 1, 2, 9, 555},
			want: []any{uint16(0), uint16(1), uint16(2), uint16(9), uint16(555)},
		},
		{
			name: "[]uint32",
			in:   []uint32{0, 1, 2, 9, 55555},
			want: []any{uint32(0), uint32(1), uint32(2), uint32(9), uint32(55555)},
		},
		{
			name: "[]uint64",
			in:   []uint64{0, 1, 2, 9, 999999999},
			want: []any{uint64(0), uint64(1), uint64(2), uint64(9), uint64(999999999)},
		},
		{
			name: "[]float32",
			in:   []float32{1, 9999999, 54321, -9},
			want: []any{float32(1), float32(9999999), float32(54321), float32(-9)},
		},
		{
			name: "[]float64",
			in:   []float64{1, 9999999, 54321, -9},
			want: []any{float64(1), float64(9999999), float64(54321), float64(-9)},
		},
		{
			name: "[]time.Time",
			in:   []time.Time{dt1, dt2, dt3},
			want: []any{dt1, dt2, dt3},
		},
		{
			name: "[]*time.Time",
			in:   []*time.Time{&dt1, &dt2, &dt3},
			want: []any{&dt1, &dt2, &dt3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := convertList(tc.in)
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestRunner_Compare(t *testing.T) {
	t.Parallel()

	r1 := map[string]any{"f1": 123, "f2": true, "f3": "hello world"}

	testCases := []struct {
		name string
		p    *runner
		want any
	}{
		{
			name: "both nil",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr1"}},
					TransformationType: enums.TransformCompare,
					CompareType:        enums.InList,
					ResultType:         enums.BooleanType,
				},
			},
		},
		{
			name: "v1 nil",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr2"}},
					TransformationType: enums.TransformCompare,
					CompareType:        enums.IsLike,
					ResultType:         enums.BooleanType,
				},
			},
		},
		{
			name: "v2 nil",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr3"}},
					TransformationType: enums.TransformCompare,
					CompareType:        enums.IsEqual,
					ResultType:         enums.BooleanType,
					InputValues:        map[int]any{1: "123"},
				},
			},
		},
		{
			name: "exists - true",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr4"}},
					TransformationType: enums.TransformCompare,
					CompareType:        enums.Exists,
					ResultType:         enums.BooleanType,
					InputValues:        map[int]any{1: "123"},
				},
			},
			want: true,
		},
		{
			name: "exists - false",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr5"}},
					TransformationType: enums.TransformCompare,
					CompareType:        enums.Exists,
					ResultType:         enums.BooleanType,
				},
			},
			want: false,
		},
		{
			name: "not exists - true",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr6"}},
					TransformationType: enums.TransformCompare,
					CompareType:        enums.NotExists,
					ResultType:         enums.BooleanType,
				},
			},
			want: true,
		},
		{
			name: "not exists - false",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr7"}},
					TransformationType: enums.TransformCompare,
					CompareType:        enums.NotExists,
					ResultType:         enums.BooleanType,
					InputValues:        map[int]any{1: "123"},
				},
			},
			want: false,
		},
		{
			name: "in list - true",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr8"}},
					TransformationType: enums.TransformCompare,
					CompareType:        enums.InList,
					ResultType:         enums.BooleanType,
					InputFields:        map[int]string{1: "f1"},
					InputValues:        map[int]any{2: []int{1, 2, 123, 999}},
				},
			},
			want: true,
		},
		{
			name: "match regex - true",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr9"}},
					TransformationType: enums.TransformCompare,
					CompareType:        enums.MatchRegex,
					Insensitive:        true,
					ResultType:         enums.BooleanType,
					InputFields:        map[int]string{1: "f3"},
					InputValues:        map[int]any{2: "hello .*"},
				},
			},
			want: true,
		},
		{
			name: "lesserThan - true",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr9"}},
					TransformationType: enums.TransformCompare,
					CompareType:        enums.IsLesser,
					ResultType:         enums.BooleanType,
					InputFields:        map[int]string{1: "f1"},
					InputValues:        map[int]any{2: "124"},
				},
			},
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f1 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: enums.IntegerType}
			f2 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: enums.BooleanType}
			f3 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: enums.StringType}

			t1 := &schema.Table{
				Object: schema.Object{
					Parent: schema.Parent{ID: "t1"},
					Fields: []*schema.Field{f1, f2, f3},
				},
				Transforms: []*schema.Transformation{tc.p.transform},
			}

			ds := &schema.Datasource{
				Parent: schema.Parent{ID: "ds"},
				Tables: []*schema.Table{t1},
			}
			ds.Fix(nil)

			got := tc.p.compare()
			assert.Equal(t, tc.want, got)
		})
	}
}
