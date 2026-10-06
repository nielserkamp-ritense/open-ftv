// Package main contains the main function for the app.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pap/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pap/server"
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

	logger.Info("git version", "hash", config.GitHash, "tag", config.GitTag)

	if cfg.DatabaseURL == "" {
		logger.Error("a Postgres URI is required to start the PAP: set PAP_DATABASE_URL or --database-url")
		os.Exit(1)
	}

	srv := server.NewService(cfg, logger)
	srv.Serve()
}
