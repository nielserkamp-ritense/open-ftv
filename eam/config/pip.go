package config

import "flag"

// PIP contains the configuration variables for a generic PIP.
type PIP struct {
	Store        string
	StoreRecurse bool
	PullConfigs  string
}

func (p *PIP) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&p.Store, "pip-store", "", "Path where PIP attribute files are stored")
	fs.BoolVar(&p.StoreRecurse, "pip-store-recurse", false, "Search PIP attribute file storage recursively")
	fs.StringVar(&p.PullConfigs, "pip-pull-configs", "", "Path where PIP pull configuration files are stored")
}
