package transforming

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

func TestRunner_Convert(t *testing.T) {
	t.Parallel()

	r1 := map[string]any{"f1": 123, "f2": true, "f3": "hello world"}

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
					TransformationType: enums.TransformConvert,
					ResultType:         enums.StringType,
				},
			},
		},
		{
			name: "integer",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr2"}},
					TransformationType: enums.TransformConvert,
					ResultType:         enums.IntegerType,
					InputValues:        map[int]any{1: "-123"},
				},
			},
			want: int64(-123),
		},
		{
			name: "unsigned integer",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr2"}},
					TransformationType: enums.TransformConvert,
					ResultType:         enums.UnsignedIntegerType,
					InputValues:        map[int]any{1: "123"},
				},
			},
			want: uint64(123),
		},
		{
			name: "float",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr2"}},
					TransformationType: enums.TransformConvert,
					ResultType:         enums.FloatType,
					InputValues:        map[int]any{1: "-123.456"},
				},
			},
			want: -123.456,
		},
		{
			name: "integer",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr2"}},
					TransformationType: enums.TransformConvert,
					ResultType:         enums.BooleanType,
					InputValues:        map[int]any{1: "1"},
				},
			},
			want: true,
		},
		{
			name: "date",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr2"}},
					TransformationType: enums.TransformConvert,
					ResultType:         enums.DateType,
					InputValues:        map[int]any{1: "2025-01-04"},
				},
			},
			want: time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "string",
			p: &runner{
				rec: r1,
				transform: &schema.Transformation{
					Object:             schema.Object{Parent: schema.Parent{ID: "tr2"}},
					TransformationType: enums.TransformConvert,
					ResultType:         enums.StringType,
					InputValues:        map[int]any{1: 123.567},
				},
			},
			want: "123.567",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.p.convert()
			assert.Equal(t, tc.want, got)
		})
	}
}
