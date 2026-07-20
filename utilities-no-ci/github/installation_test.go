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

func TestNewInstallationClient(t *testing.T) {
	d := t.TempDir()
	secretsFile := d + "/secrets.yaml"

	path1 := d + "/bad_key_file"
	fErr := os.WriteFile(path1, []byte("hello world"), 0600)
	require.NoError(t, fErr)

	key, rsaErr := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, rsaErr)
	require.NotNil(t, key)

	path2 := d + "/good_signing_key.pem"
	b := x509.MarshalPKCS1PrivateKey(key)
	b = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: b})
	fErr = os.WriteFile(path2, b, 0600)
	require.NoError(t, fErr)

	testCases := []struct {
		name           string
		appID          string
		installationID string
		secretsFile    string
		keyFile        string
		wantErr        bool
	}{
		{name: "bad secrets file", appID: "1", installationID: "2", secretsFile: "/not/a/valid/path/kemosabe!", wantErr: true},
		{name: "bad appID", appID: "xyz", installationID: "2", secretsFile: secretsFile, wantErr: true},
		{name: "bad installationID", appID: "1", installationID: "xyz", secretsFile: secretsFile, wantErr: true},
		{name: "bad key file", appID: "1", installationID: "2", secretsFile: secretsFile, keyFile: "/not/a/valid/path/popeye!", wantErr: true},
		{name: "bad key", appID: "1", installationID: "2", secretsFile: secretsFile, keyFile: path1, wantErr: true},
		{name: "all good", appID: "1", installationID: "2", secretsFile: secretsFile, keyFile: path2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if strings.HasPrefix(tc.secretsFile, d) {
				fErr = os.WriteFile(tc.secretsFile, []byte(fmt.Sprintf(template, tc.appID, "abc", tc.installationID, tc.keyFile)), 0644)
				require.NoError(t, fErr)
			}

			got, err := NewInstallationClient(context.Background(), tc.secretsFile)
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
