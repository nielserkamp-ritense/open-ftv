package x509

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCA(t *testing.T) {
	t.Run("root ca", func(t *testing.T) {
		cert := RootCA()
		assert.Equal(t, "FTV Root Authority", cert.Subject.CommonName)
		assert.Equal(t, []string{"FTV reference implementation"}, cert.Subject.Organization)
		assert.NoError(t, cert.VerifyHostname("127.0.0.1"))
	})
}
