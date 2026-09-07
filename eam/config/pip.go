package config

import (
	"context"
	"log/slog"

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
// This builds a self-authorization PIP loaded from local attribute/entity files.
// The files are the source of truth. An in-memory KV store is used as its runtime index.
func (p *PIP) NewSelfAuthzPIP(ctx context.Context, logger *slog.Logger) (*pip.PIP, error) {
	opts := []pip.Option{
		pip.WithKeyValueDB(memory.New(), ""),
		pip.WithFileStore(p.Store, p.StoreRecurse),
	}

	if p.PullConfigs != "" {
		opts = append(opts, pip.WithPullConfigs(p.PullConfigs))
	}

	return pip.New(ctx, logger, opts...)
}
