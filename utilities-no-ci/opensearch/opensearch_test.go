package opensearch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	user      = "admin"
	pswd      = "Str0ng_P@ssw0rd!"
	endpoints = []string{"https://search.localhost"}
)

func TestNewBase(t *testing.T) {
	t.Run("new base", func(t *testing.T) {
		b, err := newBase(user, pswd, endpoints)
		require.NoError(t, err)
		require.NotNil(t, b)
		assert.NotNil(t, b.client)
	})
}
