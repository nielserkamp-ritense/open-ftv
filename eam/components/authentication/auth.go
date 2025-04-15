package authentication

import (
	"context"
	"log/slog"
	"os"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// Authenticator represents the interface to authenticate users.
type Authenticator interface {
	AuthenticateUser(ctx context.Context, user, pswd string) error // user must exist and password must match.
	AuthenticateApiKey(ctx context.Context, apikey string) error   // api-key must exist.
}

func newBase(opts []Option) *base {
	b := &base{
		ctx:      context.Background(),
		log:      slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		entities: models.NewEntitySet(),
	}

	for i := range opts {
		opts[i](b)
	}
	return b
}

type base struct {
	ctx      context.Context
	log      *slog.Logger
	entities models.EntitySet
}
