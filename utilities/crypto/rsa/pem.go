// Package rsa contains functionality to handle private and/or public RSA keys.
package rsa

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	pem2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/crypto/pem"
)

// LoadPrivateKeyFile returns the private key from the given PEM block, or an error if it fails.
func LoadPrivateKeyFile(path string) (*rsa.PrivateKey, error) {
	b, err := pem2.LoadFile(path)
	if err != nil {
		return nil, err
	}
	return PrivateKeyFromPEM(b)
}

// PrivateKeyFromPEM returns the private key from the given PEM block, or an error if it fails.
func PrivateKeyFromPEM(b *pem.Block) (*rsa.PrivateKey, error) {
	if b == nil {
		return nil, errors.New("cannot parse empty private key")
	}
	if b.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("unsupported key type: %s", b.Type)
	}
	return x509.ParsePKCS1PrivateKey(b.Bytes)
}
