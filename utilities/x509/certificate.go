package x509

import (
	"crypto/rsa"
	"crypto/x509"
	"math/big"
	"net"
)

// NewCert creates a new x509 client certificate and returns it in PEM encoded format, along with the private key.
func NewCert(cfg *Config) ([]byte, *rsa.PrivateKey, error) {
	cfg.fix()

	template := x509.Certificate{
		SerialNumber:       big.NewInt(int64(cfg.Serial)),
		Subject:            cfg.Name,
		Issuer:             cfg.CA.Issuer,
		SignatureAlgorithm: x509.SHA256WithRSA,
		PublicKeyAlgorithm: x509.RSA,
		Version:            3,
		IPAddresses:        []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:          cfg.ValidFrom,
		NotAfter:           cfg.ValidTo,
		SubjectKeyId:       []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:           x509.KeyUsageDigitalSignature,
	}

	return generate(&template, cfg.CA, cfg.CAKey, cfg.KeySize)
}
