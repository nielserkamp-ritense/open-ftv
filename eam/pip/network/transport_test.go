package network

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequest_transport_reloadsRotatedClientCertificate(t *testing.T) {
	t.Parallel()

	_, caCert, caKey := makeIntermediate(t)
	certFile, keyFile := makeCert(t, caCert, caKey)

	req := &Request{CertFile: certFile, KeyFile: keyFile}

	rt, err := req.transport()
	require.NoError(t, err)

	getCert := rt.(*http.Transport).TLSClientConfig.GetClientCertificate

	before, err := getCert(nil)
	require.NoError(t, err)

	rotatedCert, rotatedKey := makeCert(t, caCert, caKey)
	require.NoError(t, os.Rename(rotatedCert, certFile))
	require.NoError(t, os.Rename(rotatedKey, keyFile))

	after, err := getCert(nil)
	require.NoError(t, err)

	assert.NotEqual(t, before.Certificate[0], after.Certificate[0])
}

func TestRequest_transport_retriesAfterFailure(t *testing.T) {
	t.Parallel()

	_, caCert, caKey := makeIntermediate(t)
	certFile, keyFile := makeCert(t, caCert, caKey)

	dir := t.TempDir()
	req := &Request{CertFile: filepath.Join(dir, "cert.pem"), KeyFile: filepath.Join(dir, "key.pem")}

	_, err := req.transport()
	require.Error(t, err)

	require.NoError(t, os.Rename(certFile, req.CertFile))
	require.NoError(t, os.Rename(keyFile, req.KeyFile))

	_, err = req.transport()
	require.NoError(t, err)
}
