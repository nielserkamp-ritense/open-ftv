//go:build external

package github

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadSecrets(t *testing.T) {
	testCases := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{name: "bad path", path: "", wantErr: true},
		{name: "bad content", path: "../../testdata/github/bad.yaml", wantErr: true},
		{name: "good", path: "../../testdata/github/github1.yaml"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := loadSecrets(tc.path)
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

var template = `
---
appID: %s
clientID: %s
installationID: %s
keyFile: %s
`
