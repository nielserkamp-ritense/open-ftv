// Package main contains the main function for a Policy Decision Point using AuthZEN.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/server"
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

	if cfg.ADLService == "" {
		logger.Error("a service name for the Authorization Decision Log is required to start the PDP: set PDP_ADL_SERVICE_NAME or --adl-service-name")
		os.Exit(1)
	}

	if cfg.ADLDatabaseURL == "" {
		logger.Error("a Postgres URI for the Authorization Decision Log is required to start the PDP: set PDP_ADL_PG_URL or --adl-pg-url")
		os.Exit(1)
	}

	srv := server.NewService(cfg, logger)
	if err := srv.Serve(); err != nil {
		os.Exit(1) // the service that failed has logged why.
	}
}
