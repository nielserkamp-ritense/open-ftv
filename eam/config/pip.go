package config

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

// PIP contains the configuration variables for a generic PIP.
type PIP struct {
	Store        string `json:"pipStore,omitempty"   yaml:"pip.store.path,omitempty"      env:"PIP_STORE"         flag:"pip-store"         desc:"Path where PIP attribute files are stored"`
	StoreRecurse bool   `json:"pipRecurse,omitempty" yaml:"pip.store.recurse,omitempty"   env:"PIP_STORE_RECURSE" flag:"pip-store-recurse" desc:"Search PIP attribute file storage recursively"`
	PullConfigs  string `json:"pipPullCfg,omitempty" yaml:"pip.pull.configPath,omitempty" env:"PIP_PULL_CONFIGS"  flag:"pip-pull-configs"  desc:"Path where PIP pull configuration files are stored"`
}

// NewSelfAuthzPIP instantiates a new PIP using the given configuration.
//
// This builds a self-authorization PIP loaded from local (bundled) attribute/entity files, if
// any: the files are the source of truth, so an in-memory KV store is used as its runtime
// index. This is not a durable persistence backend and must not be used for customer-managed
// attribute/entity data.
func (p *PIP) NewSelfAuthzPIP(ctx context.Context, logger *slog.Logger, language models.Language) (*pip.PIP, error) {
	opts := []pip.Option{
		pip.WithKeyValueDB(memory.New(), ""),
		pip.WithFileStore(p.Store, p.StoreRecurse),
	}

	if p.PullConfigs != "" {
		opts = append(opts, pip.WithPullConfigs(p.PullConfigs))
	}

	return pip.New(ctx, logger, opts...)
}
