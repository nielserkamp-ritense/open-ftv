package pip

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
)

// See RFC-6750 for the OAuth2 Authorization bearer scheme!
func (p *pip) processAuth(req *components.Request, auth string, a models.AttributeSet) {
	if strings.HasPrefix(auth, "Bearer ") {
		p.processBearer(req, auth[7:], a)
	} else {
		p.logger.Warn("unsupported authorization header", "request-uid", req.UID, "authorization", auth)
	}
}

func (p *pip) processBearer(req *components.Request, bearer string, a models.AttributeSet) {
	token, err := jwt.Parse(bearer, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("secret1!"), nil
	})
	if err != nil {
		p.logger.Error("failed to parse jwt token", "request-uid", req.UID, "error", err)
		return
	}

	m := map[string]any{
		models.AttrValid:   token.Valid,
		models.AttrHeaders: token.Header,
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		m[models.AttrClaims] = claims
	} else {
		p.logger.Warn("jwt token without claims", "request-uid", req.UID, "jwt", token)
	}

	a.AddAttribute(models.AttrJWT, m)
}
