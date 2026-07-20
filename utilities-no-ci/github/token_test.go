//go:build external

package github

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	before60 := now.Add(-60 * time.Second)
	after120 := now.Add(120 * time.Second)

	d := t.TempDir()
	secretFile := d + "secrets.yaml"

	path1 := d + "/bad_key_file"
	fErr := os.WriteFile(path1, []byte("hello world"), 0600)
	require.NoError(t, fErr)

	k1, ecdsaErr := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, ecdsaErr)
	require.NotNil(t, k1)

	path2 := d + "/bad_signing_key.pem"
	b, mErr := x509.MarshalECPrivateKey(k1)
	require.NoError(t, mErr)
	b = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	fErr = os.WriteFile(path2, b, 0600)
	require.NoError(t, fErr)

	k2, rsaErr := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, rsaErr)
	require.NotNil(t, k2)

	path3 := d + "/good_signing_key.pem"
	b = x509.MarshalPKCS1PrivateKey(k2)
	b = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: b})
	fErr = os.WriteFile(path3, b, 0600)
	require.NoError(t, fErr)

	testCases := []struct {
		name       string
		path       string
		appID      int
		clientID   string
		keyFile    string
		start      time.Time
		end        time.Time
		wantErr    bool
		wantIssuer string
		wantStart  time.Time
		wantEnd    time.Time
	}{
		{
			name:    "bad path",
			path:    "/this/is/not/a/valid/path/mickey!",
			wantErr: true,
		},
		{
			name:     "bad key file",
			path:     secretFile,
			appID:    1234567,
			clientID: "aBcDeF",
			keyFile:  path1,
			wantErr:  true,
		},
		{
			name:     "bad key",
			path:     secretFile,
			appID:    1234567,
			clientID: "aBcDeF",
			keyFile:  path2,
			start:    before60,
			end:      after120,
			wantErr:  true,
		},
		{
			name:       "end before start",
			path:       secretFile,
			appID:      1234567,
			clientID:   "aBcDeF",
			keyFile:    path3,
			start:      after120,
			end:        before60,
			wantErr:    false,
			wantIssuer: "aBcDeF",
			wantStart:  after120,
			wantEnd:    after120.Add(10 * time.Minute),
		},
		{
			name:       "all good",
			path:       secretFile,
			appID:      1234567,
			clientID:   "aBcDeF",
			keyFile:    path3,
			start:      before60,
			end:        after120,
			wantIssuer: "aBcDeF",
			wantStart:  before60,
			wantEnd:    after120,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if strings.HasPrefix(tc.path, d) {
				fErr = os.WriteFile(tc.path, []byte(fmt.Sprintf(template, fmt.Sprintf("%d", tc.appID), tc.clientID, "2", tc.keyFile)), 0644)
				require.NoError(t, fErr)
			}

			got, err := GenerateToken(tc.path, tc.start, tc.end)
			if tc.wantErr {
				require.Error(t, err)
				require.Empty(t, got)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, got)

				token, err2 := jwt.Parse(
					got,
					func(token *jwt.Token) (interface{}, error) {
						if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
							return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
						}
						return &k2.PublicKey, nil
					},
					jwt.WithIssuer(tc.wantIssuer),
					jwt.WithExpirationRequired(),
				)
				require.NoError(t, err2)
				require.NotNil(t, token)

				var issuer string
				issuer, err = token.Claims.GetIssuer()
				require.NoError(t, err)
				assert.Equal(t, tc.wantIssuer, issuer)

				var nd *jwt.NumericDate
				nd, err = token.Claims.GetIssuedAt()
				require.NoError(t, err)
				assert.Zero(t, tc.wantStart.Sub(nd.Time))

				nd, err = token.Claims.GetExpirationTime()
				require.NoError(t, err)
				assert.Zero(t, tc.wantEnd.Sub(nd.Time))
			}
		})
	}
}
