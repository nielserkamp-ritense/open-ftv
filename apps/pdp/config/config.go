// Package config handles the configurable options for the PDP.
package config

import (
	"flag"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/envflag"
)

var (
	// GitHash and GitTag identify the build; overridden at build time via -ldflags -X.
	GitHash = "unknown"
	GitTag  = "unknown"
)

// Config represents our configuration variables.
type Config struct {
	config2.ServerApp
	config2.InternalServer
	config2.HealthServer
	config2.PDP
	config2.PIP
	config2.PAP
	config2.Cerbos
	config2.Authentication
	config2.Authorization
	ADLService     string
	ADLDatabaseURL string
}

// Load reads the configuration from the command-line args, falling back to PDP_ environment variables.
func Load(args []string, lookupEnv func(string) (string, bool)) (*Config, error) {
	cfg := &Config{}

	fs := flag.NewFlagSet("pdp", flag.ContinueOnError)
	cfg.Server.RegisterFlags(fs)
	cfg.Log.RegisterFlags(fs)
	cfg.InternalServer.RegisterFlags(fs)
	cfg.HealthServer.RegisterFlags(fs)
	cfg.PDP.RegisterFlags(fs)
	cfg.PIP.RegisterFlags(fs)
	cfg.PAP.RegisterFlags(fs)
	cfg.Cerbos.RegisterFlags(fs)
	cfg.Authentication.RegisterFlags(fs)
	cfg.Authorization.RegisterFlags(fs)
	fs.StringVar(&cfg.ADLService, "adl-service-name", "", "Name of this service in the Authorization Decision Log")
	fs.StringVar(&cfg.ADLDatabaseURL, "adl-pg-url", "", "Postgres URI of the Authorization Decision Log, written by the PDP and read by the Manager")

	if err := envflag.Parse(fs, args, "PDP_", lookupEnv); err != nil {
		return nil, err
	}

	if cfg.InternalHost == "" {
		cfg.InternalHost = cfg.Host
	}

	if cfg.HealthHost == "" {
		cfg.HealthHost = cfg.Host
	}

	return cfg, nil
}
