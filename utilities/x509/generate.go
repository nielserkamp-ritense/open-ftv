package x509

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func generate(cert, parent *x509.Certificate, parentKey *rsa.PrivateKey, keyBits int) ([]byte, *rsa.PrivateKey, error) {
	if keyBits == 0 {
		keyBits = 2048
	}

	key, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return nil, nil, err
	}

	if parentKey == nil {
		parentKey = key
	}

	certBytes, err2 := x509.CreateCertificate(rand.Reader, cert, parent, &key.PublicKey, parentKey)
	if err2 != nil {
		return nil, nil, err2
	}

	pemBytes := &bytes.Buffer{}
	if err = pem.Encode(pemBytes, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return nil, nil, err
	}

	return pemBytes.Bytes(), key, nil
}
