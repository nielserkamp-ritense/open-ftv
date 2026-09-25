// Package config handles the configurable options for the EAM manager.
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

// Config represents the full set of configuration variables.
type Config struct {
	config2.ServerApp
	config2.InternalServer
	config2.HealthServer
	config2.PAP
	config2.PIP
	config2.Cerbos
	config2.Authentication
	config2.Authorization
	config2.Bundle
	config2.OIDC
	DatabaseURL    string
	ADLDatabaseURL string
}

// Load reads the configuration from the command-line args, falling back to MANAGER_ environment variables.
func Load(args []string, lookupEnv func(string) (string, bool)) (*Config, error) {
	cfg := &Config{}

	fs := flag.NewFlagSet("manager", flag.ContinueOnError)
	cfg.Server.RegisterFlags(fs)
	cfg.Log.RegisterFlags(fs)
	cfg.InternalServer.RegisterFlags(fs)
	cfg.HealthServer.RegisterFlags(fs)
	cfg.PAP.RegisterFlags(fs)
	cfg.PIP.RegisterFlags(fs)
	cfg.Cerbos.RegisterFlags(fs)
	cfg.Authentication.RegisterFlags(fs)
	cfg.Authorization.RegisterFlags(fs)
	cfg.Bundle.RegisterFlags(fs)
	cfg.OIDC.RegisterFlags(fs)
	fs.StringVar(&cfg.DatabaseURL, "database-url", "", "Postgres URI of the Manager's own database (policies, attributes, entities and settings)")
	fs.StringVar(&cfg.ADLDatabaseURL, "adl-pg-url", "", "Postgres URI of the Authorization Decision Log, written by the PDP and read by the Manager")

	if err := envflag.Parse(fs, args, "MANAGER_", lookupEnv); err != nil {
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
