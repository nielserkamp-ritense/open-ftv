package x509

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// CertFromPEM returns a certificate from the given PEM encoded data.
func CertFromPEM(data []byte) (*x509.Certificate, error) {
	b, _ := pem.Decode(data)
	if b == nil || b.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("PEM data not a valid certificate")
	}
	return x509.ParseCertificate(b.Bytes)
}
