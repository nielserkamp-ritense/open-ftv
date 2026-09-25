package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/dataspaces/rvig"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/generic/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/generic/server"
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

	logger, err := slog2.Init(cfg.Output, cfg.Format, cfg.Level, cfg.Source)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to initialize logger:", err)
		os.Exit(1)
	}

	logger.Info("configuration loaded successfully", "config", cfg)

	srv, err := server.NewService(cfg, logger, rvig.FS)
	if err != nil {
		logger.Error("failed to start service", "error", err)
		os.Exit(1)
	}

	srv.Serve()
}
