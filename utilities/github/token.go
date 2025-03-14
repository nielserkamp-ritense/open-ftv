package github

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/golang-jwt/jwt/v5"

	rsa2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/crypto/rsa"
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
		end = time.Now().UTC().Add(10 * time.Minute)
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

func loadSecrets(path string) (*secrets, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read seecrets file: %w", err)
	}
	defer f.Close()

	s := secrets{}
	if err = yaml.NewDecoder(f).Decode(&s); err != nil {
		return nil, fmt.Errorf("failed to parse secrets file: %w", err)
	}

	return &s, nil
}

type secrets struct {
	AppID    string `yaml:"appID,omitempty"`
	ClientID string `yaml:"clientID,omitempty"`
	KeyFile  string `yaml:"keyFile,omitempty"`
}
