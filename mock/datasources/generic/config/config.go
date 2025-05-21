package config

import (
	"fmt"
	"log/slog"

	"gitlab.com/gjuyn/go-config/config"
	"gitlab.com/gjuyn/go-config/config-ext/yaml"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/config"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

type Config struct {
	config2.ServerApp
	DataPath string `yaml:"data.path,omitempty" env:"DATA_PATH" flag:"data" desc:"Base path for data and metadata"`
}

func New(opts ...config.Option) (*Config, *slog.Logger, error) {
	cfg := &Config{}

	// put fixed and custom configuration options in the appropriate order.
	opts = append(
		append(
			[]config.Option{
				yaml.FilesYAML(files...),
				config.EnvironmentPrefix(envPrefix),
				config.AppName(AppName),
			},
			opts...,
		),
		config.NoHelpOnError(),
	)

	_ = config.LoadConfig(cfg, opts...)

	logger, err := slog2.Init(cfg.Log.Output, cfg.Log.Format, cfg.Log.Level, cfg.Log.Source)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	if err = config.LoadConfig(cfg, append(opts, config.NoDefaults())...); err != nil {
		return nil, logger, fmt.Errorf("failed to load configuration: %w", err)
	}

	return cfg, logger, nil
}

const (
	AppName   = "Generic Datasource 0.1"
	envPrefix = "GEN_DS_"
	cfg1      = "/etc/gen-ds/default.conf"
	cfg2      = "./etc/gen-ds.yaml"
	cfg3      = "./gen-ds.yaml"
)

var files = []string{cfg1, cfg2, cfg3}
