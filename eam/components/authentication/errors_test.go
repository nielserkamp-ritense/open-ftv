package authentication

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrUnauthenticated_Error(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "nil",
			want: "authentication failure: <nil>",
		},
		{
			name: "not authorized",
			err:  fmt.Errorf("not authorized"),
			want: "authentication failure: not authorized",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := &ErrUnauthenticated{err: tc.err}
			assert.Equal(t, tc.want, e.Error())
		})
	}
}
