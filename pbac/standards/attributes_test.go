package standards

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		want string
	}{
		{name: "action", in: AttrAction, want: "action"},
		{name: "method", in: AttrMethod, want: "method"},
		{name: "doelbinding", in: AttrDoelbinding, want: "doelbinding"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.in
			assert.Equal(t, tc.want, got)
		})
	}
}
