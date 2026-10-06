package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnyToInt64(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   any
		want int64
	}{
		{name: "struct{}", in: struct{}{}},
		{name: "string 0", in: "0"},
		{name: "string 987654", in: "987654", want: 987654},
		{name: "string haha", in: "haha"},
		{name: "int64 -9999", in: int64(-9999), want: -9999},
		{name: "int -123", in: -123, want: -123},
		{name: "uint64 9999", in: uint64(9999), want: 9999},
		{name: "uint 123", in: uint(123), want: 123},
		{name: "bool true", in: true, want: 1},
		{name: "bool false", in: false},
		{name: "double 9.99", in: 9.99, want: 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := AnyToInt64(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAnyToUint64(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   any
		want uint64
	}{
		{name: "struct{}", in: struct{}{}},
		{name: "string 0", in: "0"},
		{name: "string 987654", in: "987654", want: 987654},
		{name: "string haha", in: "haha"},
		{name: "int64 9999", in: int64(9999), want: 9999},
		{name: "int 123", in: 123, want: 123},
		{name: "uint64 9999", in: uint64(9999), want: 9999},
		{name: "uint 123", in: uint(123), want: 123},
		{name: "bool true", in: true, want: 1},
		{name: "bool false", in: false},
		{name: "double 9.99", in: 9.99, want: 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := AnyToUint64(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAnyToInts(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   any
		want []int64
	}{
		{name: "nil", want: []int64{}},
		{name: "single int", in: int64(3), want: []int64{3}},
		{name: "int slice", in: []int64{1, 0, 3}, want: []int64{1, 0, 3}},
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

			got := AnyToInts(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAnyToUints(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   any
		want []uint64
	}{
		{name: "nil", want: []uint64{}},
		{name: "single uint", in: uint64(3), want: []uint64{3}},
		{name: "uint slice", in: []uint64{1, 0, 3}, want: []uint64{1, 0, 3}},
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

			got := AnyToUints(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
