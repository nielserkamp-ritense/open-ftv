package io

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSupported(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		want bool
	}{
		{name: "empty"},
		{name: "bad", in: "xyz"},
		{name: "RDF", in: "application/rdf+xml", want: true},
		{name: "ODRL", in: "text/x-policy-odrl", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsSupported(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestConvertExt(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty"},
		{name: ".zip", in: ".zip"},
		{name: ".csv", in: ".csv", want: "text/csv"},
		{name: ".json", in: ".json", want: "application/json"},
		{name: ".xml", in: ".xml", want: "application/xml"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ConvertExt(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
