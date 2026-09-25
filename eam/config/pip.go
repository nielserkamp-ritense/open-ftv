package config

import "flag"

// PIP contains the configuration variables for a generic PIP.
type PIP struct {
	Store        string `env:"PIP_STORE"         flag:"pip-store"         desc:"Path where PIP attribute files are stored"`
	StoreRecurse bool   `env:"PIP_STORE_RECURSE" flag:"pip-store-recurse" desc:"Search PIP attribute file storage recursively"`
	PullConfigs  string `env:"PIP_PULL_CONFIGS"  flag:"pip-pull-configs"  desc:"Path where PIP pull configuration files are stored"`
}

func (p *PIP) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&p.Store, "pip-store", "", "Path where PIP attribute files are stored")
	fs.BoolVar(&p.StoreRecurse, "pip-store-recurse", false, "Search PIP attribute file storage recursively")
	fs.StringVar(&p.PullConfigs, "pip-pull-configs", "", "Path where PIP pull configuration files are stored")
}
