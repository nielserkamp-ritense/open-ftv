package transforming

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/enums"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
)

func TestRunner_TestParameter(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		params   map[string]any
		s        string
		want     any
		wantType enums.FieldType
	}{
		{
			name:     "not a parameter",
			params:   map[string]any{},
			s:        "postcode",
			want:     "postcode",
			wantType: enums.StringType,
		},
		{
			name:     "not found",
			params:   map[string]any{},
			s:        ":a:",
			want:     nil,
			wantType: enums.AnyType,
		},
		{
			name:     "found (1)",
			params:   map[string]any{"a": "b", "c": 123, "d": true},
			s:        ":a:",
			want:     "b",
			wantType: enums.StringType,
		},
		{
			name:     "found (2)",
			params:   map[string]any{"a": "b", "c": 123, "d": true},
			s:        ":c:",
			want:     123,
			wantType: enums.IntegerType,
		},
		{
			name:     "found (3)",
			params:   map[string]any{"a": "b", "c": 123, "d": true},
			s:        ":d:",
			want:     true,
			wantType: enums.BooleanType,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &runner{params: tc.params}

			got, gotType := r.testParameter(tc.s)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantType, gotType)
		})
	}
}

func TestRunner_TestType(t *testing.T) {
	t.Parallel()

	str1 := struct {
		a string
		b int
	}{
		a: "a",
		b: 123,
	}

	testCases := []struct {
		name     string
		params   map[string]any
		in       any
		want     any
		wantType enums.FieldType
	}{
		{
			name:     "parameter",
			params:   map[string]any{"a": uint64(123456)},
			in:       ":a:",
			want:     uint64(123456),
			wantType: enums.UnsignedIntegerType,
		},
		{
			name:     "string",
			in:       "hello world",
			want:     "hello world",
			wantType: enums.StringType,
		},
		{
			name:     "boolean",
			in:       true,
			want:     true,
			wantType: enums.BooleanType,
		},
		{
			name:     "int8",
			in:       int8(-123),
			want:     int8(-123),
			wantType: enums.IntegerType,
		},
		{
			name:     "int32",
			in:       int32(-98765),
			want:     int32(-98765),
			wantType: enums.IntegerType,
		},
		{
			name:     "uint16",
			in:       uint16(4444),
			want:     uint16(4444),
			wantType: enums.UnsignedIntegerType,
		},
		{
			name:     "float32",
			in:       float32(5.4),
			want:     float32(5.4),
			wantType: enums.FloatType,
		},
		{
			name:     "date",
			in:       time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			want:     time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			wantType: enums.DateTimeType,
		},
		{
			name:     "struct",
			in:       str1,
			want:     str1,
			wantType: enums.AnyType,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &runner{params: tc.params}
			got, gotType := r.testType(tc.in)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantType, gotType)
		})
	}
}

func TestRunner_GetTransformationAndType(t *testing.T) {
	t.Parallel()

	fixedToday()

	r1 := map[string]any{
		"persoon.f1":    123,
		"f2":            true,
		"persoon.f3":    "hello world",
		"tr3":           "123",
		"persoon.tr4":   true,
		"geboortedatum": uint64(19970321),
	}

	f1 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f1"}}, Type: enums.IntegerType}
	f2 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f2"}}, Type: enums.BooleanType}
	f3 := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "f3"}}, Type: enums.StringType}
	geboortedatum := &schema.Field{Object: schema.Object{Parent: schema.Parent{ID: "geboortedatum"}}, Type: enums.DateType}

	tr1 := &schema.Transformation{
		Object: schema.Object{Parent: schema.Parent{ID: "tr1"}},
	}

	tr2 := &schema.Transformation{
		Object:          schema.Object{Parent: schema.Parent{ID: "tr2"}},
		InputTransforms: map[int]string{1: "tr99"},
	}

	tr3 := &schema.Transformation{
		Object:      schema.Object{Parent: schema.Parent{ID: "tr3"}},
		ResultType:  enums.StringType,
		InputFields: map[int]string{1: "f1"},
	}

	tr4 := &schema.Transformation{
		Object:      schema.Object{Parent: schema.Parent{ID: "tr4"}},
		ResultType:  enums.BooleanType,
		InputFields: map[int]string{1: "f2"},
	}

	tr5 := &schema.Transformation{
		Object:             schema.Object{Parent: schema.Parent{ID: "tr5"}},
		TransformationType: enums.TransformConvert,
		ResultType:         enums.IntegerType,
		InputTransforms:    map[int]string{1: "tr3"},
	}

	tr6 := &schema.Transformation{
		Object:             schema.Object{Parent: schema.Parent{ID: "tr6"}},
		TransformationType: enums.TransformConvert,
		ResultType:         enums.StringType,
		InputTransforms:    map[int]string{1: "tr4"},
	}

	leeftijd := &schema.Transformation{
		Object:             schema.Object{Parent: schema.Parent{ID: "leeftijd"}},
		TransformationType: enums.TransformAge,
		ResultType:         enums.IntegerType,
		InputFields:        map[int]string{1: "geboortedatum"},
	}

	volwassen := &schema.Transformation{
		Object:             schema.Object{Parent: schema.Parent{ID: "volwassen"}},
		TransformationType: enums.TransformCompare,
		CompareType:        enums.IsGreaterOrEqual,
		ResultType:         enums.BooleanType,
		InputTransforms:    map[int]string{1: "leeftijd"},
		InputValues:        map[int]any{2: 18},
	}

	t1 := &schema.Table{
		Object: schema.Object{
			Parent: schema.Parent{ID: "persoon"},
			Fields: []*schema.Field{f1, f2, f3, geboortedatum},
		},
		Transforms: []*schema.Transformation{tr3, tr4, tr5, tr6, leeftijd, volwassen},
	}

	ds := &schema.Datasource{
		Parent: schema.Parent{ID: "ds"},
		Tables: []*schema.Table{t1},
	}
	ds.Fix(nil)

	testCases := []struct {
		name     string
		r        *runner
		i        int
		want     any
		wantType enums.FieldType
		wantOK   bool
	}{
		{
			name: "not found",
			r: &runner{
				data:      r1,
				transform: tr1,
			},
			i: 1,
		},
		{
			name: "bad initialization",
			r: &runner{
				data:      r1,
				transform: tr2,
			},
			i: 1,
		},
		{
			name: "found unqualified",
			r: &runner{
				data:      r1,
				transform: tr5,
			},
			i:        1,
			want:     "123",
			wantType: enums.StringType,
			wantOK:   true,
		},
		{
			name: "found qualified",
			r: &runner{
				qualified: true,
				data:      r1,
				transform: tr6,
			},
			i:        1,
			want:     true,
			wantType: enums.BooleanType,
			wantOK:   true,
		},
		{
			name: "not found; calculate",
			r: &runner{
				data:      r1,
				transform: volwassen,
			},
			i:        1,
			want:     int64(28),
			wantType: enums.IntegerType,
			wantOK:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, gotType, gotOK := tc.r.getTransformValueAndType(tc.i)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantType, gotType)
			assert.Equal(t, tc.wantOK, gotOK)
		})
	}
}
