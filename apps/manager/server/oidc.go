package server

import (
	"context"
	"errors"
	"time"

	"github.com/MicahParks/keyfunc/v3"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
)

// pepOIDCOption builds a JWKS-backed JWT validation option from the OIDC config.
// Returns (nil, nil) when OIDC is not configured (no JWKSURL).
func pepOIDCOption(ctx context.Context, o config2.OIDC) (pep.Option, error) {
	if o.JWKSURL == "" {
		return nil, nil
	}

	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	k, err := keyfunc.NewDefaultCtx(cctx, []string{o.JWKSURL})
	if err != nil {
		return nil, errors.Join(errors.New("failed to init JWKS keyfunc"), err)
	}

	rolesClaim := o.RolesClaim
	if rolesClaim == "" {
		rolesClaim = "roles"
	}

	return pep.WithJWT(pep.JWTConfig{
		Keyfunc:    k.Keyfunc,
		Issuer:     o.Issuer,
		Audience:   o.Audience,
		RolesClaim: rolesClaim,
	}), nil
}
