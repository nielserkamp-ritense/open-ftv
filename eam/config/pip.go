package config

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

// PIP contains the configuration variables for a generic PIP.
type PIP struct {
	Store        string `yaml:"pip.store.path,omitempty" env:"PIP_STORE" flag:"pip-store" desc:"Path where PIP attribute files are stored"`
	StoreRecurse bool   `yaml:"pip.store.recurse,omitempty" env:"PIP_STORE_RECURSE" flag:"pip-store-recurse" desc:"Search PIP attribute file storage recursively"`
	PullConfigs  string `yaml:"pip.pull.configPath,omitempty" env:"PIP_PULL_CONFIGS" flag:"pip-pull-configs" desc:"Path where PIP pull configuration files are stored"`
}

// NewPIP instantiates a new PIP using the given configuration.
func (p *PIP) NewPIP(ctx context.Context, logger *slog.Logger, language models.Language) (pip.PIP, error) {
	opts := []pip.Option{pip.WithFileStore(p.Store, p.StoreRecurse)}

	if p.PullConfigs != "" {
		opts = append(opts, pip.WithPullConfigs(p.PullConfigs))
	}

	if language == models.CEDAR {
		opts = append(opts, pip.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
	}

	return pip.New(ctx, logger, opts...), nil
}
