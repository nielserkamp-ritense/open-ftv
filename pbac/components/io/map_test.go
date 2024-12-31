package io

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainsAttribute(t *testing.T) {
	testCases := []struct {
		name string
		m    map[string]any
		want bool
	}{
		{name: "nil"},
		{name: "empty", m: map[string]any{}},
		{name: "not an attribute", m: map[string]any{"a": "b", "c": 4}},
		{name: "attribute", m: map[string]any{"key": "b", "value": 4}, want: true},
		{name: "attribute with surplus", m: map[string]any{"c": "d", "float": 1.0, "value": 4, "hello": "world", "key": "b"}, want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ContainsAttribute(tc.m)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestContainsEntity(t *testing.T) {
	testCases := []struct {
		name string
		m    map[string]any
		want bool
	}{
		{name: "nil"},
		{name: "empty", m: map[string]any{}},
		{name: "not an attribute", m: map[string]any{"a": "b", "c": 4}},
		{name: "entity", m: map[string]any{"type": "b", "id": 4}, want: true},
		{name: "entity with surplus", m: map[string]any{"c": "d", "float": 1.0, "id": 4, "hello": "world", "type": "b"}, want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ContainsEntity(tc.m)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestContainsRelation(t *testing.T) {
	testCases := []struct {
		name string
		m    map[string]any
		want bool
	}{
		{name: "nil"},
		{name: "empty", m: map[string]any{}},
		{name: "not an attribute", m: map[string]any{"a": "b", "c": 4}},
		{name: "relation", m: map[string]any{"subject": map[string]any{"type": "b", "id": "a"}, "predicate": "knows", "object": map[string]any{"type": "c", "id": "d"}}, want: true},
		{name: "relation with surplus", m: map[string]any{"x": "y", "predicate": "knew", "subject": map[string]any{"type": "b", "id": "a"}, "object": map[string]any{"type": "c", "id": "d"}}, want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ContainsRelation(tc.m)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestContainsJsonLD(t *testing.T) {
	j1 := `{"a": "b"}`

	j2 := `{
  "@context": {
    "name": "http://schema.org/name",
    "image": {
      "@id": "http://schema.org/image",
      "@type": "@id"
    },
    "homepage": {
      "@id": "http://schema.org/url",
      "@type": "@id"
    }
  }
}`

	j3 := `{
  "@context": "https://json-ld.org/contexts/person.jsonld",
  "name": "Manu Sporny",
  "homepage": "http://manu.sporny.org/",
  "image": "http://manu.sporny.org/images/manu.png"
}`

	j4 := `{
  "@context": {
    "name": "http://schema.org/name",
    "image": {
      "@id": "http://schema.org/image",
      "@type": "@id"
    },
    "homepage": {
      "@id": "http://schema.org/url",
      "@type": "@id"
    }
  },
  "name": "Manu Sporny",
  "homepage": "http://manu.sporny.org/",
  "image": "http://manu.sporny.org/images/manu.png"
}`

	j5 := `{
  "@context": [
    "https://json-ld.org/contexts/person.jsonld",
    "https://json-ld.org/contexts/place.jsonld",
    {"title": "http://purl.org/dc/terms/title"}
  ],
  "@graph": [{
    "http://xmlns.com/foaf/0.1/name": "Manu Sporny",
    "homepage": "http://manu.sporny.org/",
    "depiction": "http://twitter.com/account/profile_image/manusporny"
  }, {
    "title": "The Empire State Building",
    "description": "The Empire State Building is a 102-story landmark in New York City.",
    "geo": {
      "latitude": "40.75",
      "longitude": "73.98"
    }
  }]
}`

	testCases := []struct {
		name string
		data string
		want bool
	}{
		{name: "nil"},
		{name: "empty", data: "{}"},
		{name: "not json-ld", data: j1},
		{name: "json-ld (1)", data: j2, want: true},
		{name: "json-ld (2)", data: j3, want: true},
		{name: "json-ld (3)", data: j4, want: true},
		{name: "json-ld (4)", data: j5, want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var m map[string]any

			if tc.data != "" {
				err := json.Unmarshal([]byte(tc.data), &m)
				require.NoError(t, err)
			}

			got := ContainsJsonLD(m)
			assert.Equal(t, tc.want, got)
		})
	}
}
