// Package x509 contains functionality for working with X509 certificates.
package x509

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func generate(template, parent *x509.Certificate, parentKey *rsa.PrivateKey, keySize int) (certPEM []byte, key *rsa.PrivateKey, err error) {
	if keySize == 0 {
		keySize = 256
	}

	key, err = rsa.GenerateKey(rand.Reader, keySize*8)
	if err != nil {
		return
	}

	if parentKey == nil {
		// self-signed.
		parentKey = key
	}

	var certBytes []byte
	if certBytes, err = x509.CreateCertificate(rand.Reader, template, parent, &key.PublicKey, parentKey); err != nil {
		key = nil
		return
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	return
}
