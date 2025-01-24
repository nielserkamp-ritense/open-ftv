package x509

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"sync"
	"time"
)

// RootCA returns a self-signed certificate to use as root CA.
//
// It is valid for at least 24 hours after initialization.
// Use for test-purposes only!
func RootCA() *x509.Certificate {
	rootOnce.Do(initialize)
	return rootCA
}

// RootPEM returns the PEM encoded data for the root CA.
func RootPEM() []byte {
	rootOnce.Do(initialize)
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rootCA.Raw})
}

// RootKey returns the private key for the root CA.
func RootKey() *rsa.PrivateKey {
	rootOnce.Do(initialize)
	return rootKey
}

// Roots returns a certificate pool with the root CA.
func Roots() *x509.CertPool {
	rootOnce.Do(initialize)
	return roots
}

func initialize() {
	// self-signed
	data, key, err := generate(rootTemplate, rootTemplate, nil, 0)
	if err != nil {
		panic(err)
	}

	rootKey = key
	b, _ := pem.Decode(data)
	rootCA, _ = x509.ParseCertificate(b.Bytes)

	roots = x509.NewCertPool()
	roots.AppendCertsFromPEM(data)
}

var (
	rootOnce sync.Once
	rootCA   *x509.Certificate
	rootKey  *rsa.PrivateKey
	roots    *x509.CertPool
)

var rootTemplate = &x509.Certificate{
	SerialNumber: big.NewInt(1),
	Subject: pkix.Name{
		CommonName:    "FTV Root Authority",
		Organization:  []string{"FTV reference implementation"},
		Country:       []string{"NL"},
		Locality:      []string{"Den Haag"},
		StreetAddress: []string{"Nassaulaan 12"},
		PostalCode:    []string{"2514 JS"},
	},
	SignatureAlgorithm:    x509.SHA256WithRSA,
	PublicKeyAlgorithm:    x509.RSA,
	Version:               3,
	IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	NotBefore:             time.Now().UTC().Truncate(24 * time.Hour),
	NotAfter:              time.Now().UTC().Add(48 * time.Hour),
	IsCA:                  true,
	ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	BasicConstraintsValid: true,
}
