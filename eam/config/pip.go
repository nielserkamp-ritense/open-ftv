package config

// PIP contains the configuration variables for a generic PIP.
type PIP struct {
	Store        string `env:"PIP_STORE"         flag:"pip-store"         desc:"Path where PIP attribute files are stored"`
	StoreRecurse bool   `env:"PIP_STORE_RECURSE" flag:"pip-store-recurse" desc:"Search PIP attribute file storage recursively"`
	PullConfigs  string `env:"PIP_PULL_CONFIGS"  flag:"pip-pull-configs"  desc:"Path where PIP pull configuration files are stored"`
}
