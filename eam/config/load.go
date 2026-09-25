package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"gitlab.com/gjuyn/go-config/config"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// Printer represents the interface to send a sanitized copy of the configuration to the given logger.
type Printer interface {
	LogSanitized(logger *slog.Logger)
}

// Load initializes the given configuration from defaults, the environment and the command-line options.
//
// The given configuration parameter must be a pointer to a struct, and log must point to the Log settings inside it:
// the returned logger is configured according to those settings.
//
// This function will panic if any of the processing fails.
func Load(cfg any, log *Log, opts ...config.Option) *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	fail := func(msg string, err error) {
		logger.Error(msg, "error", err)
		panic(fmt.Sprintf("%s: %s", msg, err.Error()))
	}

	// disable help messages.
	opts = append(opts, config.NoHelpOnError())

	// perform initial loading to prefetch logger variables, ignoring any errors.
	_ = config.LoadConfig(cfg, opts...)

	// initialize the logger.
	var err error
	if logger, err = slog2.Init(log.Output, log.Format, log.Level, log.Source); err != nil {
		fail("failed to initialize logger", err)
	}

	// perform second loading for error checking!
	if err = config.LoadConfig(cfg, append(opts, config.NoDefaults())...); err != nil {
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
