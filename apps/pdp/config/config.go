// Package config handles the configurable options for the PDP.
package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"gitlab.com/gjuyn/go-config/config"
	"gitlab.com/gjuyn/go-config/config-ext/yaml"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

const (
	// AppName defines the name and version of this application.
	AppName   = "OpenFTV PDP 1.0"
	envPrefix = "PDP_"
	cfg1      = "/etc/pdp/config.yaml"
	cfg2      = "./etc/pdp.yaml"
	cfg3      = "./pdp.yaml"
)

// New instantiates a new configuration from one or more files, environment variables and command-line flags.
func New(opts ...config.Option) (*Config, *slog.Logger) {
	cfg := &Config{}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	fail := func(msg string, err error) {
		logger.Error(msg, "error", err)
		panic(fmt.Errorf("%s: %w", msg, err).Error())
	}

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

	// perform initial loading to prefetch logger variables, ignoring the error.
	_ = config.LoadConfig(cfg, opts...)

	// initialize the logger.
	if l, err := util.Init(cfg.Log.Output, cfg.Log.Format, cfg.Log.Level, cfg.Log.Source); err != nil {
		fail("failed to initialize logger", err)
	} else {
		logger = l
	}

	// perform second loading for error checking!
	if err := config.LoadConfig(cfg, append(opts, config.NoDefaults())...); err != nil {
		fail("failed to load configuration", err)
	}

	if logger.Enabled(context.TODO(), slog.LevelInfo) {
		sanitized := *cfg
		sanitized.OpenSearch = *sanitized.OpenSearch.Sanitized()
		sanitized.Cerbos = *sanitized.Cerbos.Sanitized()
		logger.Info("configuration loaded successfully", "config", sanitized)
	}

	return cfg, logger
}

// Config represents our configuration variables.
type Config struct {
	config2.ServerApp
	config2.PDP
	config2.PIP
	config2.PAP
	config2.Cerbos
	config2.OpenSearch
	config2.Authentication
	config2.Authorization
}
