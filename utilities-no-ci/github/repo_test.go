//go:build external

package github

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	d := t.TempDir()

	k, rsaErr := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, rsaErr)
	require.NotNil(t, k)

	path1 := d + "/private.pem"
	b := x509.MarshalPKCS1PrivateKey(k)
	b = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: b})
	fErr := os.WriteFile(path1, b, 0600)
	require.NoError(t, fErr)

	path2 := d + "/secrets.yaml"

	testCases := []struct {
		name           string
		appID          string
		clientID       string
		installationID string
		secretsFile    string
		keyFile        string
		owner          string
		repo           string
		wantErr        bool
	}{
		{
			name:           "bad secrets file",
			appID:          "1",
			clientID:       "abc",
			installationID: "2",
			secretsFile:    "/not/a/valid/path/my/dear/watson!",
			keyFile:        path1,
			owner:          "sherlock",
			repo:           "repo",
			wantErr:        true,
		},
		{
			name:           "bad appID",
			appID:          "x",
			clientID:       "abc",
			installationID: "2",
			secretsFile:    path2,
			keyFile:        path1,
			owner:          "sherlock",
			repo:           "repo",
			wantErr:        true,
		},
		{
			name:           "bad repo",
			appID:          "1",
			clientID:       "xyz",
			installationID: "2",
			secretsFile:    path2,
			keyFile:        path1,
			owner:          "sherlock",
			repo:           "repo",
			wantErr:        true,
		},
		{
			name:        "all good",
			secretsFile: "../../bin/etc/github.yaml",
			keyFile:     "../../bin/certs/juynapp-1.pem",
			owner:       "github3t",
			repo:        "gjuijn",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if strings.HasPrefix(tc.secretsFile, d) {
				fErr = os.WriteFile(path2, []byte(fmt.Sprintf(template, tc.appID, tc.clientID, tc.installationID, tc.keyFile)), 0644)
				require.NoError(t, fErr)
			}

			got, err := New(context.Background(), tc.secretsFile, tc.owner, tc.repo)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
			}
		})
	}
}
