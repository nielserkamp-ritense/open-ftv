package x509

import (
	"crypto/rsa"
	"crypto/x509"
	"math/big"
	"net"
)

// NewCA creates a x509 intermediate CA certificate and returns it in PEM encoded format, along with its generated private key.
func NewCA(cfg *Config) ([]byte, *rsa.PrivateKey, error) {
	cfg.fix()

	out := x509.Certificate{
		SerialNumber:          big.NewInt(int64(cfg.Serial)),
		Subject:               cfg.Name,
		Issuer:                cfg.CA.Issuer,
		SignatureAlgorithm:    x509.SHA256WithRSA,
		PublicKeyAlgorithm:    x509.RSA,
		Version:               3,
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:             cfg.ValidFrom,
		NotAfter:              cfg.ValidTo,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	return generate(&out, cfg.CA, cfg.CAKey, cfg.KeySize*8)
}
