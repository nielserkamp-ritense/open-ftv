package github

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/jferrl/go-githubauth"
	"golang.org/x/oauth2"
)

// NewInstallationClient instantiates an authenticating http client for use with the GitHub API.
func NewInstallationClient(ctx context.Context, secretsPath string) (*http.Client, error) {
	s, err := loadSecrets(secretsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load secrets: %w", err)
	}

	var appID int64
	if appID, err = strconv.ParseInt(s.AppID, 10, 64); err != nil {
		return nil, fmt.Errorf("invalid appID '%s': %w", s.AppID, err)
	}

	var installationID int64
	if installationID, err = strconv.ParseInt(s.InstallationID, 10, 64); err != nil {
		return nil, fmt.Errorf("invalid appID '%s': %w", s.InstallationID, err)
	}

	var d []byte
	if d, err = os.ReadFile(s.KeyFile); err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	appTokenSource, err2 := githubauth.NewApplicationTokenSource(appID, d)
	if err2 != nil {
		return nil, fmt.Errorf("failed to create application token: %w", err2)
	}

	installationTokenSource := githubauth.NewInstallationTokenSource(installationID, appTokenSource)

	return oauth2.NewClient(ctx, installationTokenSource), nil
}
