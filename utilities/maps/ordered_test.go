package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessOrdered(t *testing.T) {
	testCases := []struct {
		name       string
		in         map[string]any
		wantKeys   []string
		wantValues []any
	}{
		{
			name: "nil",
		},
		{
			name: "empty",
			in:   map[string]any{},
		},
		{
			name:       "single",
			in:         map[string]any{"one": "hello world"},
			wantKeys:   []string{"one"},
			wantValues: []any{"hello world"},
		},
		{
			name:       "few",
			in:         map[string]any{"one": "hello world", "int": 123, "bool": true},
			wantKeys:   []string{"bool", "int", "one"},
			wantValues: []any{true, 123, "hello world"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var keys []string
			var values []any

			ProcessOrdered(tc.in, func(k string, v any) {
				keys = append(keys, k)
				values = append(values, v)
			})

			assert.Equal(t, tc.wantKeys, keys)
			assert.Equal(t, tc.wantValues, values)
		})
	}
}
