package pep

import (
	"encoding/base64"
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

func (c *collector) processBearer(bearer string) {
	if c.jwt == nil || c.jwt.Keyfunc == nil {
		c.logger.Warn("bearer token received but JWT validation is not configured")
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
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.logger.Warn("jwt token without claims", "jwt", token)
		return
	}

	c.parc.Context.AddAttributeKV(models.AttrJWT, map[string]any{
		models.AttrValid:   token.Valid,
		models.AttrHeaders: token.Header,
		models.AttrClaims:  claims,
	})

	c.mapJWTPrincipal(claims)
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
