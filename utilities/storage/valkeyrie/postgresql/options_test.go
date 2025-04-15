package postgresql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		opts        []Option
		wantMaxLife time.Duration
		wantMaxConn int32
	}{
		{
			name: "none",
		},
		{
			name:        "max life",
			opts:        []Option{WithMaxLifetime(3 * time.Minute)},
			wantMaxLife: 3 * time.Minute,
		},
		{
			name:        "max conn",
			opts:        []Option{WithMaxConnections(15)},
			wantMaxConn: 15,
		},
		{
			name:        "all",
			opts:        []Option{WithMaxConnections(10), WithMaxLifetime(2 * time.Second)},
			wantMaxLife: 2 * time.Second,
			wantMaxConn: 10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &pool{}

			for i := range tc.opts {
				tc.opts[i](p)
			}

			assert.Equal(t, tc.wantMaxLife, p.maxLife)
			assert.Equal(t, tc.wantMaxConn, p.maxConn)
		})
	}
}
