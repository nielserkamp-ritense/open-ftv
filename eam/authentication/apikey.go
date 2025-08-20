package authentication

import (
	"context"
	"fmt"
)

// AuthenticateApiKey implements the Authenticator interface.
func (b *base) AuthenticateApiKey(ctx context.Context, apikey string) error {
	app, _, _ := b.getEntity("app::" + apikey)
	if app == nil {
		return &ErrUnauthenticated{err: fmt.Errorf("api-key not found")}
	}
	return nil
}
