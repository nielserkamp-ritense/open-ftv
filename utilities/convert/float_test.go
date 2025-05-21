package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnyToFloat64(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   any
		want float64
	}{
		{name: "struct{}", in: struct{}{}},
		{name: "string 0", in: "0"},
		{name: "string 9876.54", in: "9876.54", want: 9876.54},
		{name: "string haha", in: "haha"},
		{name: "int64 9999", in: int64(9999), want: 9999},
		{name: "int 123", in: 123, want: 123},
		{name: "uint64 9999", in: uint64(9999), want: 9999},
		{name: "uint 123", in: uint(123), want: 123},
		{name: "bool true", in: true, want: 1},
		{name: "bool false", in: false},
		{name: "double 9.99", in: 9.99, want: 9.99},
		{name: "float 9.5", in: float32(9.5), want: 9.5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := AnyToFloat64(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAnyToFloats(t *testing.T) {
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

			got := AnyToFloats(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
