package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"gitlab.com/gjuyn/go-config/config"
)

// Logger represents the interface to initialize a logger.
type Logger interface {
	MakeLogger() (*slog.Logger, error)
}

// Printer represents the interface to send a sanitized copy of the configuration to the given logger.
type Printer interface {
	LogSanitized(logger *slog.Logger)
}

// Load initializes the given configuration from defaults, config files, the environment and the command-line options.
//
// The given configuration parameter must be a pointer to a struct.
//
// If the given configuration does not support initializing a logger by implementing the Logger interface,
// this function will return a standard slog.Logger printing in JSON format to STDERR.
//
// However, if you embed the Log structure from this module into your configuration structure, e.g.:
//
//	type myConfig struct {
//	  config.Log
//	}
//
// this function will return a standard slog.Logger configured according to the settings in the config.Log structure.
//
// This function will panic if any of the processing fails.
func Load(cfg any, opts ...config.Option) *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	fail := func(msg string, err error) {
		logger.Error(msg, "error", err)
		panic(fmt.Sprintf("%s: %s", msg, err.Error()))
	}

	// disable help messages.
	opts = append(opts, config.NoHelpOnError())

	// perform initial loading to prefetch logger variables, ignoring any errors.
	_ = config.LoadConfig(cfg, opts...)

	if l, ok := cfg.(Logger); ok {
		// initialize the logger.
		var err error
		if logger, err = l.MakeLogger(); err != nil {
			fail("failed to initialize logger", err)
		}
	}

	// perform second loading for error checking!
	if err := config.LoadConfig(cfg, append(opts, config.NoDefaults())...); err != nil {
		fail("failed to load configuration", err)
	}

	if logger.Enabled(context.TODO(), slog.LevelInfo) {
		if p, ok := cfg.(Printer); ok {
			// log the sanitized configuration.
			p.LogSanitized(logger)
		}
	}

	return logger
}
