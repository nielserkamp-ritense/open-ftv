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
