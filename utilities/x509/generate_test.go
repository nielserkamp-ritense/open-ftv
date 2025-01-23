package x509

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	RootCA()

	key1, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	testCases := []struct {
		name    string
		cert    *x509.Certificate
		parent  *x509.Certificate
		key     *rsa.PrivateKey
		bits    int
		wantErr bool
	}{
		{
			name:    "invalid bits",
			cert:    rootTemplate,
			parent:  rootTemplate,
			bits:    -1,
			wantErr: true,
		},
		{
			name:    "invalid key",
			cert:    rootTemplate,
			parent:  RootCA(),
			key:     key1,
			wantErr: true,
		},
		{
			name:   "good",
			cert:   rootTemplate,
			parent: rootTemplate,
			bits:   4096,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, key, err2 := generate(tc.cert, tc.parent, tc.key, tc.bits)
			if tc.wantErr {
				require.Error(t, err2)
				require.Nil(t, key)
				require.Nil(t, data)
			} else {
				require.NoError(t, err2)
				require.NotNil(t, key)
				require.NotNil(t, data)

				// TODO: ...

			}
		})
	}
}
