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

func (c *collector) processBearer(bearer string) {
	token, err := jwt.Parse(bearer, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("secret1!"), nil
	})
	if err != nil {
		c.logger.Error("failed to parse jwt token", "error", err)
		return
	}

	m := map[string]any{
		models.AttrValid:   token.Valid,
		models.AttrHeaders: token.Header,
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		m[models.AttrClaims] = claims
	} else {
		c.logger.Warn("jwt token without claims", "jwt", token)
	}

	c.parc.Context.AddAttributeKV(models.AttrJWT, m)
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
