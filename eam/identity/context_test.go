package identity

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	_, ok := FromContext(ctx)
	assert.False(t, ok, "no principal attached yet")

	p := NewPrincipal(KindUser, "alice")
	ctx = WithContext(ctx, p)

	got, ok := FromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, p, got)
}
