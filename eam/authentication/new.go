package authentication

import (
	"context"
	"log/slog"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// New instantiates the authenticator for the given type (bcrypt, or none).
func New(ctx context.Context, logger *slog.Logger, kind string, getter models.GetEntity) (Authenticator, error) {
	opts := []Option{
		WithContext(ctx),
		WithLogger(logger),
		WithEntityGetter(getter),
	}

	switch strings.ToLower(kind) {
	case "bcrypt":
		return NewBCrypt(opts...), nil
	case "none", "":
		// !!NOTE!! default is no authentication.
		return NewDummy(opts...), nil
	// case "basicauth":
	// 	return NewBasicAuth(opts...), nil
	// case "jwt":
	// 	return NewJWT(opts...), nil
	default:
		return nil, nil
	}
}
