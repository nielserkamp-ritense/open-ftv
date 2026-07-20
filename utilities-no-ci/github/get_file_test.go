//go:build external

package github

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRepo_GetFile(t *testing.T) {
	chErr := os.Chdir("../../bin/etc")
	require.NoError(t, chErr)

	r, err := New(context.Background(), "./github.yaml", "github3t", "gjuijn")
	require.NoError(t, err)
	require.NotNil(t, r)

	list, err2 := r.List("")
	require.NoError(t, err2)
	require.NotNil(t, list)

	testCases := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "bad url", url: "\000\001", wantErr: true},
		{name: "invalid url", url: "http://localhost:2345/haha/bad", wantErr: true},
		{name: "good url", url: list[0]},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err3 := r.GetFile(tc.url, 10*time.Second)
			if tc.wantErr {
				require.Error(t, err3)
				require.Nil(t, b)
			} else {
				require.NoError(t, err3)
				require.NotNil(t, b)
			}
		})
	}
}
