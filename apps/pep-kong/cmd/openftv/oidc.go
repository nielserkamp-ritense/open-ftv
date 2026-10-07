package main

import (
	"context"
	"errors"
	"time"

	"github.com/MicahParks/keyfunc/v3"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
)

const jwksFetchTimeout = 10 * time.Second

// pepOptionFromConfig builds a JWKS-backed JWT validation option from the plugin
// config. Returns (nil, nil) when OIDC is not configured (no JWKSURL).
func pepOptionFromConfig(c *Config) (pep.Option, error) {
	if c.JWKSURL == "" {
		return nil, nil
	}

	// The context also ends the goroutine that refreshes the keys, so it must outlive this call.
	k, err := keyfunc.NewDefaultOverrideCtx(
		context.Background(), []string{c.JWKSURL}, keyfunc.Override{HTTPTimeout: jwksFetchTimeout},
	)
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
