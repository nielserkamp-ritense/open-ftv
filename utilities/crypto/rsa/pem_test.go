package rsa

import (
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/require"

	pem2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/crypto/pem"
)

func TestPrivateKeyFromPEM(t *testing.T) {
	t.Parallel()

	b1 := &pem.Block{
		Type:  "bad type",
		Bytes: []byte{'a', 'b', 'c'},
	}

	b2, err := pem2.LoadFile("../../../testdata/crypto/key.pem")
	require.NoError(t, err)
	require.NotNil(t, b2)

	testCases := []struct {
		name    string
		b       *pem.Block
		wantErr bool
	}{
		{name: "nil", wantErr: true},
		{name: "bad type", b: b1, wantErr: true},
		{name: "good", b: b2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := PrivateKeyFromPEM(tc.b)
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

func TestLoadPrivateKeyFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{name: "bad", path: "/this/is/not/a/valid/path", wantErr: true},
		{name: "good", path: "../../../testdata/crypto/key.pem"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := LoadPrivateKeyFile(tc.path)
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
