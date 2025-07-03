// Package network contains support for retrieving attributes, entities and relations from external sources.
package network

import (
	"fmt"
	"os"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/pelletier/go-toml/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
)

// LoadConfig loads the external sources configuration from the given path.
func LoadConfig(path string) (*Config, error) {
	mime := io.SnifFile(path)

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	cfg := new(Config)
	switch mime {
	case io.MimeTypeYAML:
		err = yaml.NewDecoder(f).Decode(cfg)
	case io.MimeTypeJSON:
		err = json.NewDecoder(f).Decode(cfg)
	case io.MimeTypeTOML:
		err = toml.NewDecoder(f).Decode(cfg)
	default:
		err = fmt.Errorf("invalid mime-type [%s]", mime)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	cfg.prepareRequests()
	return cfg, nil
}

func (c *Config) prepareRequests() {
	for i := range c.Sources {
		source := c.Sources[i]
		for j := range source.Requests {
			source.Requests[j].prepare()
		}
	}
}

// Config defines all external sources.
type Config struct {
	Description string    `json:"description,omitempty" yaml:"description,omitempty" toml:"description,omitempty"`
	Sources     []*Source `json:"sources,omitempty" yaml:"sources,omitempty" toml:"sources,omitempty"`
}

// Source defines an external source.
type Source struct {
	Name        string     `json:"name" yaml:"name" toml:"name"`
	Description string     `json:"description,omitempty" yaml:"description,omitempty" toml:"description,omitempty"`
	Requests    []*Request `json:"requests,omitempty" yaml:"requests,omitempty" toml:"requests,omitempty"`
}
