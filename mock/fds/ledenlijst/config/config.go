package config

import (
	"fmt"
	"log/slog"
	"time"

	"gitlab.com/gjuyn/go-config/config"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

type Config struct {
	Host         string        `env:"ADDRESS,HOST" flag:"address,a" default:"0.0.0.0" desc:"Address to use for service"`
	Port         uint16        `env:"PORT" flag:"port,p" default:"8090" desc:"Port to use for https service"`
	CA           string        `env:"TLS_CA" flag:"tls-ca,ca" desc:"TLS Certificate authority to use for https service"`
	Cert         string        `env:"TLS_CERT" flag:"tls-cert,cert" desc:"TLS certificate to use for https service"`
	Key          string        `env:"TLS_KEY" flag:"tls-key,key" desc:"TLS key authority to use for https service"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" flag:"read-timeout" default:"30s" desc:"ReadEntity timeout for API requests"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" flag:"write-timeout" default:"30s" desc:"Write timeout for API requests"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT" flag:"idle-timeout" default:"300s" desc:"Idle timeout for API requests"`
	LogOutput    string        `env:"LOG_OUTPUT" flag:"log-output" default:"stdout" desc:"File for writing log data"`
	LogFormat    string        `env:"LOG_FORMAT" flag:"log-format" default:"json" desc:"Format to use when writing log data (json, text)"`
	LogLevel     string        `env:"LOG_LEVEL" flag:"log-level" default:"info" desc:"Level for writing log data (debug, info, warn, error)"`
	LogSource    bool          `env:"LOG_SOURCE" flag:"log-source" default:"true" desc:"Include source-location when writing log data"`
}

func New(opts ...config.Option) (*Config, *slog.Logger, error) {
	cfg := &Config{}

	// put fixed and custom configuration options in appropriate order.
	opts = append(
		append(
			[]config.Option{
				config.EnvironmentPrefix(envPrefix),
				config.AppName(AppName),
			},
			opts...,
		),
		config.NoHelpOnError(),
	)

	_ = config.LoadConfig(cfg, opts...)

	logger, err := slog2.Init(cfg.LogOutput, cfg.LogFormat, cfg.LogLevel, cfg.LogSource)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	if err = config.LoadConfig(cfg, append(opts, config.NoDefaults())...); err != nil {
		return nil, logger, fmt.Errorf("failed to load configuration: %w", err)
	}

	return cfg, logger, nil
}

const (
	AppName   = "FDS Ledenlijst 0.1"
	envPrefix = "FDS_LEDENLIJST_"
)
