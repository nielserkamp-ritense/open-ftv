package compare

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringsEqual(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		l1   []string
		l2   []string
		want bool
	}{
		{name: "both empty", want: true},
		{name: "first empty", l2: []string{"a", "b", "c"}},
		{name: "second empty", l1: []string{"a", "b", "c"}},
		{name: "unmatched length (1)", l1: []string{"a", "c"}, l2: []string{"a", "b", "c"}},
		{name: "unmatched length (2)", l1: []string{"a", "b", "c"}, l2: []string{"b", "c"}},
		{name: "unmatched items", l1: []string{"a", "b", "c"}, l2: []string{"c", "a", "d"}},
		{name: "matched", l1: []string{"a", "b", "c"}, l2: []string{"b", "c", "a"}, want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := StringsEqual(tc.l1, tc.l2)
			assert.Equal(t, tc.want, got)
		})
	}
}
