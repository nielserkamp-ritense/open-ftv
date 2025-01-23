package convert

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
)

func TestOpaqueString(t *testing.T) {
	s1 := ""
	s2 := "hello world"

	testCases := []struct {
		name string
		in   *string
		want string
	}{
		{name: "nil"},
		{name: "empty", in: &s1},
		{name: "hello world", in: &s2, want: s2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := OpaqueString(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestRemoveHeaderParameters(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty"},
		{name: "no charset", in: "application/xml", want: "application/xml"},
		{name: "charset", in: "application/xml; charset=utf-8", want: "application/xml"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := RemoveHeaderParameters(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAnyToString(t *testing.T) {
	testCases := []struct {
		name string
		in   any
		want string
	}{
		{name: "string", in: "hello world", want: "hello world"},
		{name: "bool true", in: true, want: "true"},
		{name: "bool false", in: false, want: "false"},
		{name: "int 0", in: 0, want: "0"},
		{name: "int -9", in: -9, want: "-9"},
		{name: "int 12345678", in: 12345678, want: "12345678"},
		{name: "uint 999", in: uint(999), want: "999"},
		{name: "int64 987654321", in: int64(-987654321), want: "-987654321"},
		{name: "double 0.0", in: 0.0, want: "0"},
		{name: "double -99.88", in: -99.88, want: "-99.88"},
		{name: "json.Number 0", in: json.Number("0"), want: "0"},
		{name: "json.Number -9876", in: json.Number("-9876"), want: "-9876"},
		{name: "json.Number 0.01234", in: json.Number("0.01234"), want: "0.01234"},
		{name: "struct{}", in: struct{}{}, want: "{}"},
		{name: "[]uint8", in: []uint8{1, 2, 3}, want: "[1 2 3]"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := AnyToString(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
