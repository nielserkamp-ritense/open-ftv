package pep

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func (c *collector) processFSC(auth string) {
	if strings.HasPrefix(auth, "Bearer ") {
		c.processFSCBearer(auth[7:])
	}
	c.logger.Warn("unsupported authorization header", "authorization", auth)
}

func (c *collector) processFSCBearer(bearer string) {
	token, err := jwt.Parse(
		bearer,
		func(token *jwt.Token) (interface{}, error) {
			if key, ok := token.Header["x5t#S256"].(string); ok {
				key2, err2 := base64.RawURLEncoding.DecodeString(key)
				if err2 == nil {
					return x509.ParsePKIXPublicKey(key2)
				}
				return nil, err2
			}
			return nil, fmt.Errorf("no key found in token")
		},
	)

	// NOTE: FSC seems to generate an unverifiable token, so we skip validation (for now).
	// It means this JWT is flagged as NOT VALID, but we can access the data that we need,
	// and it can only affect the outcome of the authorisation;
	// e.g. the authorisation may fail if the JWT claims are invalid, which is what we want anyway.
	// if a black-hat is capable of changing the FSC parameters, they would also be able to change the answer
	// from the authorisation process, so all bets are off in that case anyway.
	if err != nil && c.debug {
		c.logger.Warn("failed to parse FSC token", "error", err)
	}
	if token == nil {
		return
	}

	m := map[string]any{
		models.AttrValid:   token.Valid,
		models.AttrHeaders: token.Header,
	}

	if token.Claims != nil {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			m[models.AttrClaims] = map[string]any(claims)

			aud, ok2 := claims["aud"].(string)
			svc, ok3 := claims["svc"].(string)
			if ok2 && ok3 {
				c.newURI = fmt.Sprintf("%s/%s", aud, svc)
			}
		}
	}

	c.parc.Context.AddAttribute(models.AttrFSC, m)
}
