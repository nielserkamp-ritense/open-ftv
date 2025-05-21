package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnyToBool(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   any
		want bool
	}{
		{name: "struct{}", in: struct{}{}},
		{name: "string empty", in: ""},
		{name: "string TRUE", in: "TRUE", want: true},
		{name: "string False", in: "False"},
		{name: "string 1", in: "1", want: true},
		{name: "string 0", in: "0"},
		{name: "string abc", in: "abc"},
		{name: "string 999", in: "999"},
		{name: "bool true", in: true, want: true},
		{name: "bool false", in: false},
		{name: "int 0", in: 0},
		{name: "int 1", in: 1, want: true},
		{name: "int -1", in: -1, want: true},
		{name: "int -999", in: -999, want: true},
		{name: "int64 0", in: int64(0)},
		{name: "int64 9", in: int64(9), want: true},
		{name: "double 0", in: 0.0},
		{name: "double 0.000001", in: 0.000001, want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := AnyToBool(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAnyToBools(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   any
		want []bool
	}{
		{name: "nil", want: []bool{}},
		{name: "single bool", in: true, want: []bool{true}},
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

			got := AnyToBools(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
