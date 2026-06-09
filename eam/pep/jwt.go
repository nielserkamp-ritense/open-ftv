package pep

import (
	"github.com/golang-jwt/jwt/v5"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// JWTConfig configures OIDC bearer-token validation for the PEP. When nil on the
// PEP, bearer tokens are not trusted and yield no principal (fail closed).
type JWTConfig struct {
	// Keyfunc resolves the verification key for a token (e.g. JWKS-backed).
	Keyfunc jwt.Keyfunc
	// Issuer, when set, is required to match the token "iss" claim.
	Issuer string
	// Audience, when set, is required to be present in the token "aud" claim.
	Audience string
	// RolesClaim is the claim that holds the principal's roles. Defaults to "roles".
	RolesClaim string
}

// Option configures a PEP at construction time.
type Option func(*PEP)

// WithJWT enables OIDC bearer-token validation with the given configuration.
func WithJWT(cfg JWTConfig) Option {
	return func(p *PEP) {
		if cfg.RolesClaim == "" {
			cfg.RolesClaim = "roles"
		}
		p.jwt = &cfg
	}
}

// mapJWTPrincipal maps validated claims onto the principal: sub -> user::<sub>,
// and the roles claim -> principal "roles" attribute.
func (c *collector) mapJWTPrincipal(claims jwt.MapClaims) {
	sub, _ := claims["sub"].(string)
	if sub == "" {
		c.logger.Warn("jwt token without sub claim")
		return
	}

	c.parc.Context.AddAttributeKV(models.AttrPrincipal, PrincipalUser+"::"+sub)

	if roles := extractRoles(claims[c.jwt.RolesClaim]); len(roles) > 0 {
		c.parc.Principal.Attributes().AddAttributeKV(models.AttrRoles, roles)
	}
}

// extractRoles normalizes a roles claim value into a []string.
func extractRoles(v any) []string {
	switch t := v.(type) {
	case string:
		if t == "" {
			return nil
		}
		return []string{t}
	case []string:
		return t
	case []any:
		var out []string
		for _, e := range t {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
