// Package config handles the configurable options for the PDP.
package config

import (
	"log/slog"
	"os"

	"gitlab.com/gjuyn/go-config/config"
	"gitlab.com/gjuyn/go-config/config-ext/yaml"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
)

const (
	// AppName defines the name and version of this application.
	AppName   = "OpenFTV PDP 2.0"
	GitHash   = "unknown"
	GitTag    = "unknown"
	envPrefix = "PDP_"
	cfg1      = "/etc/pdp/config.yaml"
	cfg2      = "./etc/pdp.yaml"
	cfg3      = "./pdp.yaml"
)

// New instantiates a new configuration from one or more files, environment variables and command-line flags.
func New(opts ...config.Option) (*Config, *slog.Logger) {
	cfg := &Config{}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// put fixed and custom configuration options in appropriate order.
	opts = append(
		append(
			[]config.Option{
				yaml.FilesYAML(cfg1, cfg2, cfg3),
				config.EnvironmentPrefix(envPrefix),
				config.AppName(AppName),
			},
			opts...,
		),
		config.NoHelpOnError(),
	)

	logger = config2.Load(cfg, opts...)

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
	logger.Info(AppName)

	sanitized := *c
	sanitized.OpenSearch = *sanitized.OpenSearch.Sanitized()
	sanitized.Cerbos = *sanitized.Cerbos.Sanitized()
	sanitized.DecisionLog = *sanitized.DecisionLog.Sanitized()
	logger.Info("configuration loaded successfully", "config", sanitized)
	logger.Info("git version", "hash", GitHash, "tag", GitTag)
}

// Config represents our configuration variables.
type Config struct {
	config2.ServerApp
	config2.InternalServer
	config2.HealthServer
	config2.PDP
	config2.PIP
	config2.PAP
	config2.Cerbos
	config2.OpenSearch
	config2.Authentication
	config2.Authorization
	config2.DecisionLog
	config2.Migration
}
