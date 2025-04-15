package x509

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCA(t *testing.T) {
	t.Parallel()

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
