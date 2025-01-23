package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnyToBool(t *testing.T) {
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
			got := AnyToBool(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
