package github

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	rsa2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/crypto/rsa"
)

// GenerateToken generates a JWT token for use with GitHub API's.
func GenerateToken(secretsPath string, start, end time.Time) (string, error) {
	s, err := loadSecrets(secretsPath)
	if err != nil {
		return "", err
	}

	var key *rsa.PrivateKey
	if key, err = rsa2.LoadPrivateKeyFile(s.KeyFile); err != nil {
		return "", err
	}

	if end.Before(start) {
		end = start.Add(10 * time.Minute)
	}

	t := jwt.NewWithClaims(
		jwt.SigningMethodRS256,
		jwt.RegisteredClaims{
			Issuer:    s.ClientID,
			IssuedAt:  jwt.NewNumericDate(start),
			ExpiresAt: jwt.NewNumericDate(end),
		},
	)

	var signed string
	if signed, err = t.SignedString(key); err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signed, nil
}
