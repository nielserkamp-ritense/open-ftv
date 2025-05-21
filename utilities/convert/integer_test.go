package convert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrependZero10(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		n    uint
		size uint
		want string
	}{
		{n: 0, size: 0, want: "0"},
		{n: 1, size: 0, want: "1"},
		{n: 9, size: 0, want: "9"},
		{n: 10, size: 0, want: "10"},
		{n: 0, size: 4, want: "0000"},
		{n: 1, size: 4, want: "0001"},
		{n: 9, size: 4, want: "0009"},
		{n: 10, size: 4, want: "0010"},
		{n: 99, size: 4, want: "0099"},
		{n: 100, size: 4, want: "0100"},
		{n: 999, size: 4, want: "0999"},
		{n: 1000, size: 4, want: "1000"},
		{n: 9999, size: 4, want: "9999"},
		{n: 10000, size: 4, want: "10000"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("value %d size %d", tc.n, tc.size), func(t *testing.T) {
			t.Parallel()

			got := PrependZero10(tc.n, tc.size)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestMustInt_10(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		want  int
	}{
		{input: " x "},
		{input: ""},
		{input: "  \t\t  "},
		{input: "0"},
		{input: "1", want: 1},
		{input: "\t\t1  ", want: 1},
		{input: "  1\t", want: 1},
		{input: "9", want: 9},
		{input: "10", want: 10},
		{input: "0000"},
		{input: "0001", want: 1},
		{input: "0010", want: 10},
		{input: "0100", want: 100},
		{input: "1000", want: 1000},
		{input: "10000", want: 10000},
		{input: "0009", want: 9},
		{input: "0099", want: 99},
		{input: "0999", want: 999},
		{input: "9999", want: 9999},
		{input: "99999", want: 99999},
		{input: "-1", want: -1},
		{input: "-11", want: -11},
		{input: "-1111", want: -1111},
		{input: "  -11111  ", want: -11111},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got := MustIntDecimal(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFromInt_10(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		want  int
		fail  bool
	}{
		{input: " x ", fail: true},
		{input: "", fail: true},
		{input: "  \t\t  ", fail: true},
		{input: "0"},
		{input: "1", want: 1},
		{input: "\t\t1  ", want: 1},
		{input: "  1\t", want: 1},
		{input: "9", want: 9},
		{input: "10", want: 10},
		{input: "0000"},
		{input: "0001", want: 1},
		{input: "0010", want: 10},
		{input: "0100", want: 100},
		{input: "1000", want: 1000},
		{input: "10000", want: 10000},
		{input: "0009", want: 9},
		{input: "0099", want: 99},
		{input: "0999", want: 999},
		{input: "9999", want: 9999},
		{input: "99999", want: 99999},
		{input: "-1", want: -1},
		{input: "-11", want: -11},
		{input: "-1111", want: -1111},
		{input: "  -11111  ", want: -11111},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got, err := FromIntDecimal(tc.input)
			if tc.fail {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestMustUint_10(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		want  uint
	}{
		{input: " x "},
		{input: ""},
		{input: "  \t\t  "},
		{input: "0"},
		{input: "1", want: 1},
		{input: "\t\t1  ", want: 1},
		{input: "  1\t", want: 1},
		{input: "9", want: 9},
		{input: "10", want: 10},
		{input: "0000"},
		{input: "0001", want: 1},
		{input: "0010", want: 10},
		{input: "0100", want: 100},
		{input: "1000", want: 1000},
		{input: "10000", want: 10000},
		{input: "0009", want: 9},
		{input: "0099", want: 99},
		{input: "0999", want: 999},
		{input: "9999", want: 9999},
		{input: "99999", want: 99999},
		{input: "-1"},
		{input: "-11"},
		{input: "-1111"},
		{input: "  -11111  "},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got := MustUintDecimal(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFromUint_10(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		want  uint
		fail  bool
	}{
		{input: " x ", fail: true},
		{input: "", fail: true},
		{input: "  \t\t  ", fail: true},
		{input: "0"},
		{input: "1", want: 1},
		{input: "\t\t1  ", want: 1},
		{input: "  1\t", want: 1},
		{input: "9", want: 9},
		{input: "10", want: 10},
		{input: "0000"},
		{input: "0001", want: 1},
		{input: "0010", want: 10},
		{input: "0100", want: 100},
		{input: "1000", want: 1000},
		{input: "10000", want: 10000},
		{input: "0009", want: 9},
		{input: "0099", want: 99},
		{input: "0999", want: 999},
		{input: "9999", want: 9999},
		{input: "99999", want: 99999},
		{input: "-1", fail: true},
		{input: "-11", fail: true},
		{input: "-1111", fail: true},
		{input: "  -11111  ", fail: true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got, err := FromUintDecimal(tc.input)
			if tc.fail {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestMustUint8_10(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		want  uint8
	}{
		{input: " x "},
		{input: ""},
		{input: "  \t\t  "},
		{input: "0"},
		{input: "1", want: 1},
		{input: "\t\t1  ", want: 1},
		{input: "  1\t", want: 1},
		{input: "9", want: 9},
		{input: "10", want: 10},
		{input: "0000"},
		{input: "0001", want: 1},
		{input: "0010", want: 10},
		{input: "0100", want: 100},
		{input: "0255", want: 255},
		{input: "0256"},
		{input: "1000"},
		{input: "10000"},
		{input: "0009", want: 9},
		{input: "0099", want: 99},
		{input: "0999"},
		{input: "9999"},
		{input: "99999"},
		{input: "-1"},
		{input: "-11"},
		{input: "-1111"},
		{input: "  -11111  "},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got := MustUint8Decimal(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFromUint8_10(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		want  uint8
		fail  bool
	}{
		{input: " x ", fail: true},
		{input: "", fail: true},
		{input: "  \t\t  ", fail: true},
		{input: "0"},
		{input: "1", want: 1},
		{input: "\t\t1  ", want: 1},
		{input: "  1\t", want: 1},
		{input: "9", want: 9},
		{input: "10", want: 10},
		{input: "0000"},
		{input: "0001", want: 1},
		{input: "0010", want: 10},
		{input: "0100", want: 100},
		{input: "0255", want: 255},
		{input: "0256", fail: true},
		{input: "1000", fail: true},
		{input: "10000", fail: true},
		{input: "0009", want: 9},
		{input: "0099", want: 99},
		{input: "0999", fail: true},
		{input: "9999", fail: true},
		{input: "99999", fail: true},
		{input: "-1", fail: true},
		{input: "-11", fail: true},
		{input: "-1111", fail: true},
		{input: "  -11111  ", fail: true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got, err := FromUint8Decimal(tc.input)
			if tc.fail {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestMustUint32_10(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		want  uint32
	}{
		{input: " x "},
		{input: ""},
		{input: "  \t\t  "},
		{input: "0"},
		{input: "1", want: 1},
		{input: "\t\t1  ", want: 1},
		{input: "  1\t", want: 1},
		{input: "9", want: 9},
		{input: "10", want: 10},
		{input: "0000"},
		{input: "0001", want: 1},
		{input: "0010", want: 10},
		{input: "0100", want: 100},
		{input: "1000", want: 1000},
		{input: "10000", want: 10000},
		{input: "0009", want: 9},
		{input: "0099", want: 99},
		{input: "0999", want: 999},
		{input: "9999", want: 9999},
		{input: "99999", want: 99999},
		{input: "9999999999"},
		{input: "-1"},
		{input: "-11"},
		{input: "-1111"},
		{input: "  -11111  "},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got := MustUint32Decimal(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFromUint32_10(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		want  uint32
		fail  bool
	}{
		{input: " x ", fail: true},
		{input: "", fail: true},
		{input: "  \t\t  ", fail: true},
		{input: "0"},
		{input: "1", want: 1},
		{input: "\t\t1  ", want: 1},
		{input: "  1\t", want: 1},
		{input: "9", want: 9},
		{input: "10", want: 10},
		{input: "0000"},
		{input: "0001", want: 1},
		{input: "0010", want: 10},
		{input: "0100", want: 100},
		{input: "1000", want: 1000},
		{input: "10000", want: 10000},
		{input: "0009", want: 9},
		{input: "0099", want: 99},
		{input: "0999", want: 999},
		{input: "9999", want: 9999},
		{input: "99999", want: 99999},
		{input: "9999999999", fail: true},
		{input: "-1", fail: true},
		{input: "-11", fail: true},
		{input: "-1111", fail: true},
		{input: "  -11111  ", fail: true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got, err := FromUint32Decimal(tc.input)
			if tc.fail {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

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
