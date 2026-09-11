package pep

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// See RFC-6750 for the OAuth2 Authorization bearer scheme!
func (c *collector) processAuth(auth string) {
	switch {
	case strings.HasPrefix(auth, "Bearer "):
		c.processBearer(auth[7:])
	case strings.HasPrefix(auth, "Basic "):
		c.processBasic(auth[6:])
	default:
		c.logger.Warn("unsupported authorization header", "authorization", auth)
	}
}

// AttrError is the key under which a rejected bearer token carries the reason for the rejection
// in the jwt context attribute, next to AttrValid = false.
const AttrError = "error"

func (c *collector) processBearer(bearer string) {
	if c.jwt == nil || c.jwt.Keyfunc == nil {
		c.logger.Warn("bearer token received but JWT validation is not configured")
		c.rejectBearer("jwt validation is not configured")

		return
	}

	opts := []jwt.ParserOption{jwt.WithExpirationRequired()}
	if c.jwt.Issuer != "" {
		opts = append(opts, jwt.WithIssuer(c.jwt.Issuer))
	}
	if c.jwt.Audience != "" {
		opts = append(opts, jwt.WithAudience(c.jwt.Audience))
	}

	token, err := jwt.Parse(bearer, c.jwt.Keyfunc, opts...)
	if err != nil || !token.Valid {
		c.logger.Error("failed to validate jwt token", "error", err)
		c.rejectBearer(fmt.Sprint(err))

		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.logger.Warn("jwt token without claims", "jwt", token)
		c.rejectBearer("token without claims")

		return
	}

	c.parc.Context.AddAttributeKV(models.AttrJWT, map[string]any{
		models.AttrValid:   token.Valid,
		models.AttrHeaders: token.Header,
		// Store the claims as a plain map[string]any, not the named jwt.MapClaims type:
		// the PDP's attribute->Cedar converter matches on map[string]any, so the named
		// type would fail to convert and the whole jwt context attribute be dropped.
		models.AttrClaims: map[string]any(claims),
	})

	c.mapJWTPrincipal(claims)
}

// rejectBearer records that a bearer token was presented and refused, so the caller is not
// mistaken for one without credentials: the identity boundary answers 401, not 403.
func (c *collector) rejectBearer(reason string) {
	c.parc.Context.AddAttributeKV(models.AttrJWT, map[string]any{
		models.AttrValid: false,
		AttrError:        reason,
	})
}

func (c *collector) processBasic(basic string) {
	dec, err := base64.StdEncoding.DecodeString(basic)
	if err != nil {
		c.logger.Error("failed to decode basic authentication", "error", err)
		return
	}

	s := string(dec)

	i := strings.Index(s, ":")
	if i < 0 {
		c.logger.Error("failed to parse basic authentication", "error", "no colon character found")
		return
	}

	c.parc.Context.AddAttributeKV(models.AttrBasicUser, s[:i])
	c.parc.Context.AddAttributeKV(models.AttrBasicPswd, s[i+1:])
}
