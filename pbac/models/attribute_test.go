package models

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
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
			assert.Nil(t, got.Original())
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
			assert.Nil(t, got.Original())
			assert.Equal(t, tc.t, got.Type())

			b, err := json.Marshal(got)
			require.NoError(t, err)
			assert.Equal(t, tc.wantJSON, string(b))
		})
	}
}

func TestNewOriginalAttribute(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		value    any
		original any
		t        string
		wantYAML string
	}{
		{name: "nil", key: "nil", wantYAML: "key: nil\nvalue: null\n"},
		{name: "string", key: "s1", value: "v1", t: "string", wantYAML: "key: s1\nvalue: v1\ntype: string\n"},
		{name: "int", key: "i1", value: 123, original: "123", t: "xsd:long", wantYAML: "key: i1\nvalue: 123\noriginal: \"123\"\ntype: xsd:long\n"},
		{name: "bool", key: "b1", value: true, original: "1", t: "xsd:boolean", wantYAML: "key: b1\nvalue: true\noriginal: \"1\"\ntype: xsd:boolean\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewOriginalAttribute(tc.key, tc.value, tc.original, tc.t)
			assert.Equal(t, tc.key, got.Key())
			assert.Equal(t, tc.value, got.Value())
			assert.Equal(t, tc.original, got.Original())
			assert.Equal(t, tc.t, got.Type())

			b, err := yaml.Marshal(got)
			require.NoError(t, err)
			assert.Equal(t, tc.wantYAML, string(b))
		})
	}
}
