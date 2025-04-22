package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitKeys(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   string
		want []string
	}{
		{
			name: "empty",
		},
		{
			name: "single latin",
			in:   "some_name",
			want: []string{"some_name"},
		},
		{
			name: "single latin - double quotes",
			in:   `"some_name"`,
			want: []string{"some_name"},
		},
		{
			name: "single latin - single quotes",
			in:   `'some_name'`,
			want: []string{"some_name"},
		},
		{
			name: "multiple latin",
			in:   "some_name.hello.world.attributes",
			want: []string{"some_name", "hello", "world", "attributes"},
		},
		{
			name: "multiple latin - double quotes",
			in:   `hello."some 'name'"."with.some.dots".world`,
			want: []string{"hello", "some 'name'", "with.some.dots", "world"},
		},
		{
			name: "multiple latin - single quotes",
			in:   `hello.'some "name"'.'with.some.dots'.world`,
			want: []string{"hello", "some \"name\"", "with.some.dots", "world"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := splitKeys(tc.in)
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestFindElement(t *testing.T) {
	t.Parallel()

	m1 := map[string]any{"a": "b", "c": 1, "e": true}
	m2 := map[string]any{"a": map[string]any{"b": map[string]any{"c": 123.456}}, "c": 1, "e": true}

	testCases := []struct {
		name string
		key  string
		data any
		want any
	}{
		{name: "nil", key: "mykey"},
		{name: "not a map", key: "mykey", data: "mykey"},
		{name: "empty map", key: "mykey", data: map[string]any{}},
		{name: "no key", data: m1, want: m1},
		{name: "invalid data", key: "a.b", data: m1},
		{name: "not found", key: "mykey", data: m1},
		{name: "found - first level", key: "c", data: m1, want: 1},
		{name: "found - third level", key: "a.b.c", data: m2, want: 123.456},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := findElement(splitKeys(tc.key), tc.data)
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestCodeOrValue(t *testing.T) {
	t.Parallel()

	m1 := map[string]any{
		"hello": "world",
		"int":   123,
		"float": 999.444,
		"bool":  true,
	}

	testCases := []struct {
		name  string
		code  string
		value any
		m     map[string]any
		want  any
	}{
		{name: "both empty", m: m1},
		{name: "code empty", value: "xyz", m: m1, want: "xyz"},
		{name: "code not found", code: "xyz", value: true, m: m1, want: true},
		{name: "empty map", code: "hello", value: 555, m: map[string]any{}, want: 555},
		{name: "code found", code: "hello", value: 123, m: m1, want: "world"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := codeOrValue(tc.code, tc.value, tc.m)
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestCodeOrValueString(t *testing.T) {
	t.Parallel()

	m1 := map[string]any{
		"hello": "world",
		"int":   123,
		"float": 999.444,
		"bool":  true,
	}

	testCases := []struct {
		name  string
		code  string
		value string
		m     map[string]any
		want  string
	}{
		{name: "both empty", m: m1},
		{name: "code empty", value: "xyz", m: m1, want: "xyz"},
		{name: "code not found", code: "xyz", value: "true", m: m1, want: "true"},
		{name: "empty map", code: "hello", value: "555", m: map[string]any{}, want: "555"},
		{name: "code found", code: "bool", value: "false", m: m1, want: "true"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := codeOrValueString(tc.code, tc.value, tc.m)
			assert.EqualValues(t, tc.want, got)
		})
	}
}
