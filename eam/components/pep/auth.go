package pep

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// See RFC-6750 for the OAuth2 Authorization bearer scheme!
func (c *collector) processAuth(auth string) {
	if strings.HasPrefix(auth, "Bearer ") {
		c.processBearer(auth[7:])
	} else {
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

	c.parc.Context.AddAttribute(models.AttrJWT, m)
}
