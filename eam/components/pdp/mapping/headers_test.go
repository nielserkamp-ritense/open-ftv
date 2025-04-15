package mapping

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapper_FromHeaders(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		headers []string
		in      any
		want    string
	}{
		{name: "no headers", in: map[string]any{"x": "y"}},
		{name: "single header - not found", headers: []string{"q"}, in: map[string]string{"x": "y"}},
		{name: "single header - found", headers: []string{"X"}, in: map[string]string{"x": "y"}, want: "y"},
		{name: "multiple headers - not found", headers: []string{"x", "Y", "z"}, in: map[string]any{"q": "qq"}},
		{name: "multiple headers - found", headers: []string{"x", "Y", "z"}, in: map[string]any{"Z": "qq"}, want: "qq"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := &base{}
			m.configure([]Option{WithHeaderKeys(tc.headers...)})

			got := m.fromHeaders(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
