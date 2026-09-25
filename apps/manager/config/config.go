// Package config handles the configurable options for the EAM manager.
package config

import (
	"log/slog"
	"os"

	"gitlab.com/gjuyn/go-config/config"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
)

var (
	// GitHash and GitTag identify the build; overridden at build time via -ldflags -X.
	GitHash = "unknown"
	GitTag  = "unknown"
)

const envPrefix = "MANAGER_"

// New instantiates a new configuration from environment variables and command-line flags.
func New(opts ...config.Option) (*Config, *slog.Logger) {
	cfg := &Config{}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// put fixed and custom configuration options in appropriate order.
	opts = append(
		append(
			[]config.Option{
				config.EnvironmentPrefix(envPrefix),
			},
			opts...,
		),
		config.NoHelpOnError(),
	)

	logger = config2.Load(cfg, &cfg.Log, opts...)

	if cfg.InternalHost == "" {
		cfg.InternalHost = cfg.Host
	}

	if cfg.HealthHost == "" {
		cfg.HealthHost = cfg.Host
	}

	return cfg, logger
}

// LogSanitized implements the config.Printer interface.
func (c *Config) LogSanitized(logger *slog.Logger) {
	logger.Info("git version", "hash", GitHash, "tag", GitTag)
}

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
	DatabaseURL    string `env:"DATABASE_URL" flag:"database-url" desc:"Postgres URI of the Manager's own database (policies, attributes, entities and settings)"`
	ADLDatabaseURL string `env:"ADL_PG_URL" flag:"adl-pg-url" desc:"Postgres URI of the Authorization Decision Log, written by the PDP and read by the Manager"`
}
