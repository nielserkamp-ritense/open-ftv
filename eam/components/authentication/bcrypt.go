package authentication

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// NewBCrypt instantiates a new user authentication handler.
//
// This handler uses the given set of entities to verify users with BCrypt encoded passwords.
func NewBCrypt(opts ...Option) Authenticator {
	return &bcryptAuth{base: *newBase(opts)}
}

// AuthenticateUser implements the Authenticator interface.
func (a *bcryptAuth) AuthenticateUser(_ context.Context, user, pswd string) error {
	u := a.entities.GetEntity("user::" + user)
	if u == nil {
		return &ErrUnauthenticated{err: fmt.Errorf("user not found")}
	}

	h := convert.AnyToString(u.Attributes().GetAttributeValue("password"))
	if h == "" {
		return &ErrUnauthenticated{err: fmt.Errorf("invalid user configuration")}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(h), []byte(pswd)); err != nil {
		return &ErrUnauthenticated{err: err}
	}
	return nil
}

type bcryptAuth struct {
	base
}
