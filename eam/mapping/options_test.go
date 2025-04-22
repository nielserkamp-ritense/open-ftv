package mapping

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		opts           []Option
		wantHeaderKeys []string
	}{
		{name: "none"},
		{name: "header key", opts: []Option{WithHeaderKeys("abc")}, wantHeaderKeys: []string{"abc"}},
		{name: "header keys", opts: []Option{WithHeaderKeys("abc", "def", "ghi")}, wantHeaderKeys: []string{"abc", "def", "ghi"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := &base{}
			for i := range tc.opts {
				tc.opts[i](m)
			}

			assert.EqualValues(t, tc.wantHeaderKeys, m.headerKeys)
		})
	}
}
