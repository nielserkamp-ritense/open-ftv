package transforming

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
)

func TestDMYFromNumber(t *testing.T) {
	t.Parallel()

	fixedToday()

	testCases := []struct {
		name  string
		in    any
		wantY int
		wantM int
		wantD int
	}{
		{name: "invalid", in: "hello", wantY: 0, wantM: 0, wantD: 0},
		{name: "string", in: "20250406", wantY: 2025, wantM: 4, wantD: 6},
		{name: "integer", in: 20250507, wantY: 2025, wantM: 5, wantD: 7},
		{name: "unsigned integer", in: uint64(20250608), wantY: 2025, wantM: 6, wantD: 8},
		{name: "float64", in: 20250709.0, wantY: 2025, wantM: 7, wantD: 9},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			y, m, d := dmyFromNumber(tc.in)
			assert.Equal(t, tc.wantY, y)
			assert.Equal(t, tc.wantM, m)
			assert.Equal(t, tc.wantD, d)
		})
	}
}

func TestRunner_Age(t *testing.T) {
	t.Parallel()

	fixedToday()

	r1 := map[string]any{"f1": 123, "f2": true, "f3": "hello world"}

	dt1 := fmt.Sprintf("%.4d%.2d%.2d", 2008, 6, 18)
	dt2 := fmt.Sprintf("%.4d00%.2d", 2002, 18)
	dt3 := fmt.Sprintf("%.4d%.2d00", 2002, 7)

	testCases := []struct {
		name string
		p    *runner
		want any
	}{
		{
			name: "nil",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr1"}},
					TransformationType: enums.TransformAge,
					ResultType:         enums.IntegerType,
					IsPII:              true,
				},
			},
			want: nil,
		},
		{
			name: "invalid date",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr2"}},
					TransformationType: enums.TransformAge,
					ResultType:         enums.IntegerType,
					IsPII:              true,
					InputValues:        map[int]any{1: "18490101"},
				},
			},
			want: nil,
		},
		{
			name: "future date",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr3"}},
					TransformationType: enums.TransformAge,
					ResultType:         enums.IntegerType,
					IsPII:              true,
					InputValues:        map[int]any{1: "20251230"},
				},
			},
			want: nil,
		},
		{
			name: "string - full date",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr4"}},
					TransformationType: enums.TransformAge,
					ResultType:         enums.UnsignedIntegerType,
					IsPII:              true,
					InputValues:        map[int]any{1: dt1},
				},
			},
			want: uint64(17),
		},
		{
			name: "string - no month",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr5"}},
					TransformationType: enums.TransformAge,
					ResultType:         enums.UnsignedIntegerType,
					IsPII:              true,
					InputValues:        map[int]any{1: dt2},
				},
			},
			want: uint64(23),
		},
		{
			name: "string - no day",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr6"}},
					TransformationType: enums.TransformAge,
					ResultType:         enums.StringType,
					IsPII:              true,
					InputValues:        map[int]any{1: dt3},
				},
			},
			want: "22",
		},
		{
			name: "integer - no month, no day",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr7"}},
					TransformationType: enums.TransformAge,
					ResultType:         enums.IntegerType,
					IsPII:              true,
					InputValues:        map[int]any{1: 20000000},
				},
			},
			want: int64(25),
		},
		{
			name: "date time",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr7"}},
					TransformationType: enums.TransformAge,
					ResultType:         enums.FloatType,
					IsPII:              true,
					InputValues:        map[int]any{1: time.Date(1981, 9, 10, 0, 0, 0, 0, time.UTC)},
				},
			},
			want: 43.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.p.age()
			assert.Equal(t, tc.want, got)
		})
	}
}

func fixedToday() {
	todaySetter.Do(func() {
		today = func() time.Time { return time.Date(2025, 6, 18, 0, 0, 0, 0, time.UTC) }
	})
}

var todaySetter sync.Once
