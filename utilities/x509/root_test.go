package x509

import (
	"crypto/x509"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCA(t *testing.T) {
	t.Run("root ca", func(t *testing.T) {
		cert := RootCA()
		assert.Equal(t, "FTV Root Authority", cert.Subject.CommonName)
		assert.Equal(t, []string{"FTV reference implementation"}, cert.Subject.Organization)
		assert.NoError(t, cert.VerifyHostname("127.0.0.1"))

		data := RootPEM()
		require.NotNil(t, data)

		cert2, err := CertFromPEM(data)
		require.NoError(t, err)
		assert.Equal(t, "FTV Root Authority", cert2.Subject.CommonName)
		assert.Equal(t, []string{"FTV reference implementation"}, cert2.Subject.Organization)
		assert.NoError(t, cert2.VerifyHostname("127.0.0.1"))
	})
}

func TestRootCA_Error(t *testing.T) {
	t.Run("root ca - error", func(t *testing.T) {
		temp := rootTemplate
		rootTemplate = &x509.Certificate{SerialNumber: big.NewInt(-1)}

		defer func() {
			e := recover()
			require.NotNil(t, e)
			rootTemplate = temp
		}()

		initialize()

		rootTemplate = temp
		t.Fail()
	})
}
