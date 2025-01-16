package models

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAttribute(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		value any
	}{
		{name: "nil", key: "nil"},
		{name: "string", key: "s1", value: "v1"},
		{name: "int", key: "i1", value: 123},
		{name: "bool", key: "b1", value: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewAttribute(tc.key, tc.value)
			assert.Equal(t, tc.key, got.Key())
			assert.Equal(t, tc.value, got.Value())
			assert.Empty(t, got.Type())
		})
	}
}

func TestNewAttributeWithType(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		value    any
		t        string
		wantJSON string
	}{
		{name: "nil", key: "nil", wantJSON: `{"key":"nil","value":null}`},
		{name: "string", key: "s1", value: "v1", t: "string", wantJSON: `{"key":"s1","value":"v1","type":"string"}`},
		{name: "int", key: "i1", value: 123, t: "xsd:long", wantJSON: `{"key":"i1","value":123,"type":"xsd:long"}`},
		{name: "bool", key: "b1", value: true, t: "xsd:boolean", wantJSON: `{"key":"b1","value":true,"type":"xsd:boolean"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewAttributeWithType(tc.key, tc.value, tc.t)
			assert.Equal(t, tc.key, got.Key())
			assert.Equal(t, tc.value, got.Value())
			assert.Equal(t, tc.t, got.Type())

			b, err := json.Marshal(got)
			require.NoError(t, err)
			assert.Equal(t, tc.wantJSON, string(b))
		})
	}
}
