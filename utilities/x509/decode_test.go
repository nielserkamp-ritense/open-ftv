package x509

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCertFromPEM(t *testing.T) {
	t.Parallel()

	RootCA()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	data1, err2 := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &key.PublicKey, key)
	require.NoError(t, err2)
	require.NotNil(t, data1)

	buf1 := &bytes.Buffer{}
	err = pem.Encode(buf1, &pem.Block{Type: "CERTIFICATE", Bytes: data1})
	require.NoError(t, err)

	buf2 := &bytes.Buffer{}
	err = pem.Encode(buf2, &pem.Block{Type: "PRIVATE-KEY", Bytes: data1})
	require.NoError(t, err)

	testCases := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{name: "nil", wantErr: true},
		{name: "empty", data: []byte{}, wantErr: true},
		{name: "not valid", data: []byte("**** BEGIN DATA ****\n"), wantErr: true},
		{name: "wrong type", data: buf2.Bytes(), wantErr: true},
		{name: "good", data: buf1.Bytes()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cert, err3 := CertFromPEM(tc.data)
			if tc.wantErr {
				require.Error(t, err3)
				require.Nil(t, cert)
			} else {
				require.NoError(t, err3)
				require.NotNil(t, cert)

				// TODO: ...

			}
		})
	}
}
