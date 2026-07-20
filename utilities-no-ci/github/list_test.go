//go:build external

package github

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepo_List(t *testing.T) {
	r, err := New(context.Background(), "../../bin/etc/github.yaml", "github3t", "gjuijn")
	require.NoError(t, err)
	require.NotNil(t, r)

	testCases := []struct {
		name    string
		prefix  string
		wantErr bool
	}{
		{name: "bad prefix", prefix: "../haha", wantErr: true},
		{name: "no prefix"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err2 := r.List(tc.prefix)
			if tc.wantErr {
				require.NotNil(t, err2)
				require.Nil(t, got)
			} else {
				require.NoError(t, err2)
				require.NotNil(t, got)
			}
		})
	}
}
