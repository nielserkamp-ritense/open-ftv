package network

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func makeIntermediate(t *testing.T) (caFile string, caCert *x509.Certificate, caKey *rsa.PrivateKey) {
	now := time.Now().UTC()

	rootCert, rootKey := makeSigned(t, &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               root,
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:             now.AddDate(0, 0, -2),
		NotAfter:              now.Add(96 * time.Hour),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}, nil, nil)

	validFrom := now.Truncate(24 * time.Hour)
	caCert, caKey = makeSigned(t, &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               ca,
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:             validFrom,
		NotAfter:              validFrom.Add(48 * time.Hour),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}, rootCert, rootKey)

	caFile = filepath.Join(t.TempDir(), "ca.pem")
	err := os.WriteFile(caFile, certPEM(caCert), 0o600)
	require.NoError(t, err)

	return
}

func makeCert(t *testing.T, caCert *x509.Certificate, caKey *rsa.PrivateKey) (certFile, keyFile string) {
	now := time.Now().UTC().Truncate(time.Second)

	leafCert, key := makeSigned(t, &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      cert,
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    now,
		NotAfter:     now.Add(time.Hour),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}, caCert, caKey)

	keyData := x509.MarshalPKCS1PrivateKey(key)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyData})

	d := t.TempDir()

	certFile = filepath.Join(d, "cert.pem")
	err := os.WriteFile(certFile, append(certPEM(leafCert), certPEM(caCert)...), 0o600)
	require.NoError(t, err)

	keyFile = filepath.Join(d, "key.pem")
	err = os.WriteFile(keyFile, keyPEM, 0o600)
	require.NoError(t, err)

	return
}

// makeSigned generates a key for template and signs it with parentKey, or self-signs it when parent is nil.
func makeSigned(t *testing.T, template, parent *x509.Certificate, parentKey *rsa.PrivateKey) (*x509.Certificate, *rsa.PrivateKey) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	if parent == nil {
		parent = template
		parentKey = key
	}

	der, err := x509.CreateCertificate(rand.Reader, template, parent, &key.PublicKey, parentKey)
	require.NoError(t, err)

	c, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	return c, key
}

func certPEM(c *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})
}

var root = pkix.Name{
	Country:       []string{"NL"},
	Organization:  []string{"FTV reference implementation"},
	Locality:      []string{"Den Haag"},
	StreetAddress: []string{"Nassaulaan 12"},
	PostalCode:    []string{"2514 JS"},
	CommonName:    "FTV Root Authority",
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
