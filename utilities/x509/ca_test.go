package x509

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCA(t *testing.T) {
	t.Parallel()

	RootCA()

	subject := pkix.Name{
		Country:       []string{"NL"},
		Organization:  []string{"FTV"},
		Province:      []string{"ZH"},
		StreetAddress: []string{"Mickey Mouse laan 1"},
		PostalCode:    []string{"0099ZZ"},
		CommonName:    "",
	}

	now := defaultNow()

	testCases := []struct {
		name    string
		serial  int
		keySize int
		subject pkix.Name
		ca      *x509.Certificate
		caKey   *rsa.PrivateKey
		from    time.Time
		to      time.Time
		expires time.Duration
		wantErr bool
	}{
		{
			name:    "no root, no from, no to, no expires",
			serial:  100,
			keySize: 512,
			subject: subject,
		},
		{
			name:    "root CA, no from, no to, no expires",
			serial:  100,
			subject: subject,
			ca:      RootCA(),
			caKey:   RootKey(),
		},
		{
			name:    "root CA, no from, no to, expires",
			serial:  100,
			subject: subject,
			ca:      RootCA(),
			caKey:   RootKey(),
			expires: 15 * time.Minute,
		},
		{
			name:    "root CA, from, no to, no expires",
			serial:  100,
			subject: subject,
			ca:      RootCA(),
			caKey:   RootKey(),
			from:    now,
		},
		{
			name:    "no root CA, from, no to, expires",
			serial:  100,
			subject: subject,
			from:    now.AddDate(0, 0, -1),
			expires: 25 * time.Hour,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := &Config{
				Serial:       tc.serial,
				KeySize:      tc.keySize,
				Name:         tc.subject,
				CA:           tc.ca,
				CAKey:        tc.caKey,
				ValidFrom:    tc.from,
				ExpiresAfter: tc.expires,
			}

			data, key, err := NewCA(cfg)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, data)
				assert.Nil(t, key)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, data)
				assert.NotNil(t, key)

				if tc.keySize > 0 {
					assert.Equal(t, tc.keySize, key.Size())
				} else {
					assert.Equal(t, 256, key.Size())
				}

				b, rest := pem.Decode(data)
				assert.Empty(t, rest)
				assert.NotNil(t, b)
				assert.Equal(t, "CERTIFICATE", b.Type)

				cert, err2 := x509.ParseCertificate(b.Bytes)
				require.NoError(t, err2)
				require.NotNil(t, cert)

				assert.Equal(t, tc.subject.CommonName, cert.Subject.CommonName)
				assert.Equal(t, tc.subject.Organization, cert.Subject.Organization)
				assert.Equal(t, tc.subject.Province, cert.Subject.Province)
				assert.Equal(t, tc.subject.StreetAddress, cert.Subject.StreetAddress)
				assert.Equal(t, tc.subject.PostalCode, cert.Subject.PostalCode)

				assert.GreaterOrEqual(t, cert.NotBefore, cfg.ValidFrom.Add(-defaultMargin))
				assert.LessOrEqual(t, cert.NotBefore, cfg.ValidFrom.Add(defaultMargin))
				assert.GreaterOrEqual(t, cert.NotAfter, cfg.ValidTo.Add(-defaultMargin))
				assert.LessOrEqual(t, cert.NotAfter, cfg.ValidTo.Add(defaultMargin))

				_, err3 := cert.Verify(x509.VerifyOptions{
					Roots:       Roots(),
					KeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
					CurrentTime: time.Now().UTC(),
				})
				require.NoError(t, err3)
			}
		})
	}
}
