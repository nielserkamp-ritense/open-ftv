package csv

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

func TestToSlice_bool(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   any
		want []bool
	}{
		{name: "nil", want: []bool{}},
		{name: "single bool", in: true, want: []bool{true}},
		{name: "single int", in: 5, want: []bool{true}},
		{name: "bool slice", in: []bool{true, false, true}, want: []bool{true, false, true}},
		{name: "any slice", in: []any{1, 2.1, true, "hello world"}, want: []bool{true, true, true, false}},
		{name: "int slice", in: []int{1, 0, 1}, want: []bool{true, false, true}},
		{name: "int8 slice", in: []int8{1, 0, 3}, want: []bool{true, false, false}},
		{name: "int16 slice", in: []int16{1, 0, 3}, want: []bool{true, false, false}},
		{name: "int32 slice", in: []int32{1, 0, 3}, want: []bool{true, false, false}},
		{name: "int64 slice", in: []int64{1, 0, 0}, want: []bool{true, false, false}},
		{name: "uint slice", in: []uint{1, 0, 3}, want: []bool{true, false, false}},
		{name: "uint8 slice", in: []uint8{1, 0, 3}, want: []bool{true, false, false}},
		{name: "uint16 slice", in: []uint16{1, 0, 3}, want: []bool{true, false, false}},
		{name: "uint32 slice", in: []uint32{1, 0, 3}, want: []bool{true, false, false}},
		{name: "uint64 slice", in: []uint64{1, 0, 3}, want: []bool{true, false, false}},
		{name: "string slice", in: []string{"true", "xxx", "false"}, want: []bool{true, false, false}},
		{name: "float32 slice", in: []float32{1.0, 0.0, 3.3}, want: []bool{true, false, false}},
		{name: "float64 slice", in: []float64{1, 0.0, 1}, want: []bool{true, false, true}},
		{name: "unsupported type", in: new(struct{}), want: []bool{false}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := toSlice(tc.in, convert.AnyToBool)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestToSlice_float64(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   any
		want []float64
	}{
		{name: "nil", want: []float64{}},
		{name: "single float", in: 3.21, want: []float64{3.21}},
		{name: "float slice", in: []float64{1.1, 0, 3.3}, want: []float64{1.1, 0, 3.3}},
		{name: "any slice", in: []any{1, 2.1, true, "hello world"}, want: []float64{1, 2.1, 1, 0}},
		{name: "int slice", in: []int{1, 0, 1}, want: []float64{1, 0, 1}},
		{name: "int8 slice", in: []int8{1, 0, 3}, want: []float64{1, 0, 3}},
		{name: "int16 slice", in: []int16{1, 0, 3}, want: []float64{1, 0, 3}},
		{name: "int32 slice", in: []int32{1, 0, 3}, want: []float64{1, 0, 3}},
		{name: "int64 slice", in: []int64{1, 0, 4}, want: []float64{1, 0, 4}},
		{name: "uint slice", in: []uint{1, 0, 3}, want: []float64{1, 0, 3}},
		{name: "uint8 slice", in: []uint8{1, 0, 3}, want: []float64{1, 0, 3}},
		{name: "uint16 slice", in: []uint16{1, 0, 3}, want: []float64{1, 0, 3}},
		{name: "uint32 slice", in: []uint32{1, 0, 3}, want: []float64{1, 0, 3}},
		{name: "uint64 slice", in: []uint64{1, 0, 3}, want: []float64{1, 0, 3}},
		{name: "float32 slice", in: []float32{1.0, 0.0, 3.5}, want: []float64{1.0, 0.0, 3.5}},
		{name: "float64 slice", in: []float64{1, 0.0, 1.2345}, want: []float64{1, 0.0, 1.2345}},
		{name: "bool slice", in: []bool{true, false, true}, want: []float64{1, 0, 1}},
		{name: "string slice", in: []string{"true", "xxx", "6.66"}, want: []float64{0, 0, 6.66}},
		{name: "unsupported type", in: new(struct{}), want: []float64{0}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := toSlice(tc.in, convert.AnyToFloat64)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestToSlice_int64(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   any
		want []int64
	}{
		{name: "nil", want: []int64{}},
		{name: "single int", in: int64(3), want: []int64{3}},
		{name: "single float", in: 2.6, want: []int64{3}},
		{name: "single bool", in: true, want: []int64{1}},
		{name: "int64 slice", in: []int64{1, 0, 3}, want: []int64{1, 0, 3}},
		{name: "any slice", in: []any{1, 2.1, true, "hello world"}, want: []int64{1, 2, 1, 0}},
		{name: "int slice", in: []int{1, 0, 1}, want: []int64{1, 0, 1}},
		{name: "int8 slice", in: []int8{1, 0, 3}, want: []int64{1, 0, 3}},
		{name: "int16 slice", in: []int16{1, 0, 3}, want: []int64{1, 0, 3}},
		{name: "int32 slice", in: []int32{1, 0, 3}, want: []int64{1, 0, 3}},
		{name: "uint slice", in: []uint{1, 0, 3}, want: []int64{1, 0, 3}},
		{name: "uint8 slice", in: []uint8{1, 0, 3}, want: []int64{1, 0, 3}},
		{name: "uint16 slice", in: []uint16{1, 0, 3}, want: []int64{1, 0, 3}},
		{name: "uint32 slice", in: []uint32{1, 0, 3}, want: []int64{1, 0, 3}},
		{name: "uint64 slice", in: []uint64{1, 0, 3}, want: []int64{1, 0, 3}},
		{name: "float32 slice", in: []float32{1.0, 0.0, 6}, want: []int64{1, 0, 6}},
		{name: "float64 slice", in: []float64{1, 0.0, 1.2345}, want: []int64{1, 0, 1}},
		{name: "bool slice", in: []bool{true, false, true}, want: []int64{1, 0, 1}},
		{name: "string slice", in: []string{"true", "xxx", "666"}, want: []int64{0, 0, 666}},
		{name: "unsupported type", in: new(struct{}), want: []int64{0}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := toSlice(tc.in, convert.AnyToInt64)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestToSlice_uint64(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   any
		want []uint64
	}{
		{name: "nil", want: []uint64{}},
		{name: "single uint", in: uint64(3), want: []uint64{3}},
		{name: "uint64 slice", in: []uint64{1, 0, 3}, want: []uint64{1, 0, 3}},
		{name: "any slice", in: []any{1, 2.1, true, "hello world"}, want: []uint64{1, 2, 1, 0}},
		{name: "int slice", in: []int{1, 0, 1}, want: []uint64{1, 0, 1}},
		{name: "int8 slice", in: []int8{1, 0, 3}, want: []uint64{1, 0, 3}},
		{name: "int16 slice", in: []int16{1, 0, 3}, want: []uint64{1, 0, 3}},
		{name: "int32 slice", in: []int32{1, 0, 3}, want: []uint64{1, 0, 3}},
		{name: "int64 slice", in: []int64{1, 0, 3}, want: []uint64{1, 0, 3}},
		{name: "uint slice", in: []uint{1, 0, 3}, want: []uint64{1, 0, 3}},
		{name: "uint8 slice", in: []uint8{1, 0, 3}, want: []uint64{1, 0, 3}},
		{name: "uint16 slice", in: []uint16{1, 0, 3}, want: []uint64{1, 0, 3}},
		{name: "uint32 slice", in: []uint32{1, 0, 3}, want: []uint64{1, 0, 3}},
		{name: "float32 slice", in: []float32{1.0, 0.0, 6}, want: []uint64{1, 0, 6}},
		{name: "float64 slice", in: []float64{1, 0.0, 1.2345}, want: []uint64{1, 0, 1}},
		{name: "bool slice", in: []bool{true, false, true}, want: []uint64{1, 0, 1}},
		{name: "string slice", in: []string{"true", "xxx", "666"}, want: []uint64{0, 0, 666}},
		{name: "unsupported type", in: new(struct{}), want: []uint64{0}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := toSlice(tc.in, convert.AnyToUint64)
			assert.Equal(t, tc.want, got)
		})
	}
}
