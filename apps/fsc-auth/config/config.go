// Package config handles the configurable options for the FSC Auth plugin.
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
	AppName   = "FSC-AUTH plugin 0.1"
	envPrefix = "FSC_AUTH_"
	cfg1      = "/etc/fsc-auth/default.conf"
	cfg2      = "./etc/fsc-auth.yaml"
	cfg3      = "./fsc-auth.yaml"
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
	return cfg, logger
}

// LogSanitized implements the config.Printer interface.
func (c *Config) LogSanitized(logger *slog.Logger) {
	sanitized := *c
	sanitized.Cerbos.Sanitized()
	sanitized.OpenSearch.Sanitized()
	logger.Info("configuration loaded successfully", "config", sanitized)
}

// Config represents the full set of configuration variables.
type Config struct {
	config2.ServerApp
	config2.PAP
	config2.PIP
	config2.Cerbos
	config2.OpenSearch
	RequestMappings string
}
