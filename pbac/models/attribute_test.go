package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			assert.Equal(t, tc.key, got.Key)
			assert.Equal(t, tc.value, got.Value)
		})
	}
}
