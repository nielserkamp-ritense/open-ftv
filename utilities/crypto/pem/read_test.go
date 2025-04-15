package pem

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{name: "bad path", path: "/this/is/not/a/valid/path/goofy!", wantErr: true},
		{name: "private key", path: "../../../testdata/crypto/key.pem"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := LoadFile(tc.path)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
			}
		})
	}
}

func TestLoadStream(t *testing.T) {
	t.Parallel()

	r1, err := os.Open("../../../testdata/crypto/key.pem")
	require.NoError(t, err)
	r1.Close()

	r2, err2 := os.Open("../../../testdata/crypto/key.pem")
	require.NoError(t, err2)

	testCases := []struct {
		name    string
		r       io.ReadCloser
		wantErr bool
	}{
		{name: "already closed", r: r1, wantErr: true},
		{name: "good", r: r2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defer tc.r.Close()

			got, err3 := LoadStream(tc.r)
			if tc.wantErr {
				require.Error(t, err3)
				require.Nil(t, got)
			} else {
				require.NoError(t, err3)
				require.NotNil(t, got)
			}
		})
	}
}
