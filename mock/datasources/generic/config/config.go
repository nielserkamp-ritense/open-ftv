package config

import (
	"flag"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/envflag"
)

type Config struct {
	config2.ServerApp
	DataPath string
}

// Load reads the configuration from the command-line args, falling back to GEN_DS_ environment variables.
func Load(args []string, lookupEnv func(string) (string, bool)) (*Config, error) {
	cfg := &Config{}

	fs := flag.NewFlagSet("generic-datasource", flag.ContinueOnError)
	cfg.Server.RegisterFlags(fs)
	cfg.Log.RegisterFlags(fs)
	fs.StringVar(&cfg.DataPath, "data-path", "", "Base path for data and metadata")

	if err := envflag.Parse(fs, args, "GEN_DS_", lookupEnv); err != nil {
		return nil, err
	}

	return cfg, nil
}

const AppName = "Generic Datasource 0.1"
