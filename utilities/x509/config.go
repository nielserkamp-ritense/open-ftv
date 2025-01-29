package x509

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"time"
)

// Config contains the parameters for generating a new certificate.
//
// The Serial number and the Name parameter are mandatory.
//
// Serial should be a non-zero positive number. Increasing with each renewal of a certificate.
//
// Name should be filled according to x509 requirements.
//
// KeySize should be 256, 512 or a similar valid key size.
// If it is zero, it will be set to 256, resulting in 2048 bits of entropy.
//
// Optional fields are "fixed" in the following order:
//
//  1. if CA is nil, RootCA() and RootKey() will be used;
//     note: the key must not be nil if CA is filled!
//  2. if ValidFrom is zero, it will be set to the current date & time.
//  3. if ExpiresAfter is zero, it will be set to 5 seconds.
//  4. if ValidTo is zero, it will be set to ValidFrom + ExpiresAfter.
//  5. finally, both ValidFrom and ValidTo are truncated to the seconds value;
//     e.g. the nanoseconds in the time.Time{} value are set to zero.
type Config struct {
	Serial       int               `json:"serial" yaml:"serial" toml:"serial"`
	KeySize      int               `json:"keySize" yaml:"keySize" toml:"keySize"`
	Name         pkix.Name         `json:"name" yaml:"name" toml:"name"`
	CA           *x509.Certificate `json:"ca,omitempty" yaml:"ca,omitempty" toml:"ca,omitempty"`
	CAKey        *rsa.PrivateKey   `json:"caKey,omitempty" yaml:"caKey,omitempty" toml:"caKey,omitempty"`
	ValidFrom    time.Time         `json:"validFrom" yaml:"validFrom" toml:"validFrom"`
	ValidTo      time.Time         `json:"validTo" yaml:"validTo" toml:"validTo"`
	ExpiresAfter time.Duration     `json:"expiresAfter,omitempty" yaml:"expiresAfter,omitempty" toml:"expiresAfter,omitempty"`
}

func (cfg *Config) fix() {
	if cfg.CA == nil {
		cfg.CA = RootCA()
		cfg.CAKey = RootKey()
	}

	if cfg.ValidFrom.IsZero() {
		cfg.ValidFrom = defaultNow()
	}

	if cfg.ExpiresAfter == 0 {
		cfg.ExpiresAfter = defaultExpires
	}

	if cfg.ValidTo.IsZero() {
		cfg.ValidTo = cfg.ValidFrom.Add(cfg.ExpiresAfter)
	}

	cfg.ValidFrom = cfg.ValidFrom.Truncate(defaultTruncate)
	cfg.ValidTo = cfg.ValidTo.Truncate(defaultTruncate)
}

var (
	defaultNow      = func() time.Time { return time.Now().UTC().Truncate(defaultTruncate).Add(-5 * time.Second) }
	defaultExpires  = 10 * time.Second
	defaultTruncate = time.Second
)
