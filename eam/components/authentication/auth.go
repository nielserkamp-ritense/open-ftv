package authentication

import "context"

// Authenticator represents the interface to authenticate users.
type Authenticator interface {
	Authenticate(ctx context.Context, user, pswd string) error
}
