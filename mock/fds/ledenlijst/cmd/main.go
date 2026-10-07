package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/fds/ledenlijst/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/fds/ledenlijst/server"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func main() {
	cfg, err := config.Load(os.Args[1:], os.LookupEnv)
	if errors.Is(err, flag.ErrHelp) {
		return
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load configuration:", err)
		os.Exit(1)
	}

	logger, err := slog2.Init(cfg.Log.Output, cfg.Log.Format, cfg.Log.Level, cfg.Log.Source)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to initialize logger:", err)
		os.Exit(1)
	}

	logger.Info("configuration loaded successfully", "config", cfg)

	if err := server.NewService(cfg, logger).Serve(); err != nil {
		os.Exit(1) // the service that failed has logged why.
	}
}
