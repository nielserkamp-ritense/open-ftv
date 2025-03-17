package github

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

func loadSecrets(path string) (*secrets, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets file: %w", err)
	}
	defer f.Close()

	s := secrets{}
	if err = yaml.NewDecoder(f).Decode(&s); err != nil {
		return nil, fmt.Errorf("failed to parse secrets file: %w", err)
	}

	return &s, nil
}

type secrets struct {
	AppID          string `yaml:"appID,omitempty"`
	ClientID       string `yaml:"clientID,omitempty"`
	InstallationID string `yaml:"installationID,omitempty"`
	KeyFile        string `yaml:"keyFile,omitempty"`
}
