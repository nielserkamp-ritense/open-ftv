package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaults(t *testing.T) {
	t.Parallel()

	got := Defaults()
	require.NotNil(t, got)

	assert.Equal(t, DefaultHeaderTitle, got.HeaderTitle)
	assert.Equal(t, DefaultHeaderColor, got.HeaderColor)
	assert.Equal(t, DefaultTitleColor, got.TitleColor)
	assert.Empty(t, got.Logo)
	assert.Empty(t, got.LogoMediaType)
	assert.Empty(t, got.Updated)
	assert.Empty(t, got.UpdatedBy)

	// Callers must be able to mutate the returned value without affecting later Defaults().
	got.HeaderTitle = "changed"
	assert.Equal(t, DefaultHeaderTitle, Defaults().HeaderTitle)
}
