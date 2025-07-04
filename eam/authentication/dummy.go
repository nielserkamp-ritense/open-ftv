package authentication

import (
	"context"
)

// NewDummy instantiates a new fake authentication handler.
//
// This handler authenticates anyone and anything without hesitation.
func NewDummy(opts ...Option) Authenticator {
	return &dummy{base: *newBase(opts)}
}

// AuthenticateUser implements the Authenticator interface.
func (a *dummy) AuthenticateUser(_ context.Context, _, _ string) error {
	return nil
}

type dummy struct {
	base
}
