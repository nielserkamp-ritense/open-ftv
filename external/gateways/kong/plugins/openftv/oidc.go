package main

import (
	"context"
	"errors"
	"time"

	"github.com/MicahParks/keyfunc/v3"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
)

// pepOptionFromConfig builds a JWKS-backed JWT validation option from the plugin
// config. Returns (nil, nil) when OIDC is not configured (no JWKSURL).
func pepOptionFromConfig(c *Config) (pep.Option, error) {
	if c == nil || c.JWKSURL == "" {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	k, err := keyfunc.NewDefaultCtx(ctx, []string{c.JWKSURL})
	if err != nil {
		return nil, errors.Join(errors.New("failed to initialize JWKS keyfunc"), err)
	}

	return pep.WithJWT(pep.JWTConfig{
		Keyfunc:    k.Keyfunc,
		Issuer:     c.Issuer,
		Audience:   c.Audience,
		RolesClaim: c.RolesClaim,
	}), nil
}
