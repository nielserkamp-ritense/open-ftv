package network

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	certs "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/x509"
)

func makeIntermediate(t *testing.T) (rootFile, caFile string, caCert *x509.Certificate, caKey *rsa.PrivateKey) {
	now := time.Now().UTC()

	caData, key, err := certs.NewCA(&certs.Config{
		Serial:       1,
		KeySize:      256,
		Name:         ca,
		CA:           certs.RootCA(),
		CAKey:        certs.RootKey(),
		ValidFrom:    now.Truncate(24 * time.Hour),
		ExpiresAfter: 48 * time.Hour,
	})
	require.NoError(t, err)

	d := t.TempDir()

	rootFile = filepath.Join(d, "root.pem")
	err = os.WriteFile(rootFile, certs.RootPEM(), 0644)
	require.NoError(t, err)

	caFile = filepath.Join(d, "ca.pem")
	err = os.WriteFile(caFile, caData, 0644)
	require.NoError(t, err)

	caCert, err = certs.CertFromPEM(caData)
	require.NoError(t, err)

	caKey = key
	return
}

func makeCert(t *testing.T, caCert *x509.Certificate, caKey *rsa.PrivateKey) (certFile, keyFile string) {
	now := time.Now().UTC()

	certData, key, err := certs.NewCert(&certs.Config{
		Serial:       1,
		KeySize:      256,
		Name:         cert,
		CA:           caCert,
		CAKey:        caKey,
		ValidFrom:    now,
		ExpiresAfter: time.Hour,
	})
	require.NoError(t, err)

	keyData := x509.MarshalPKCS1PrivateKey(key)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyData})

	caData := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCert.Raw})

	d := t.TempDir()

	certFile = filepath.Join(d, "cert.pem")
	err = os.WriteFile(certFile, append(certData, caData...), 0644)
	require.NoError(t, err)

	keyFile = filepath.Join(d, "key.pem")
	err = os.WriteFile(keyFile, keyPEM, 0600)
	require.NoError(t, err)

	return
}

var ca = pkix.Name{
	Country:      []string{"NL"},
	Organization: []string{"VNGR"},
	Locality:     []string{"Den Haag"},
	CommonName:   "VNGR",
}

var cert = pkix.Name{
	Country:      []string{"NL"},
	Organization: []string{"FDS"},
	Locality:     []string{"Den Haag"},
	CommonName:   "127.0.0.1",
}
