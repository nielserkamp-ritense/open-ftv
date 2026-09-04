// Package main contains the main function for a Policy Decision Point using AuthZEN.
package main

import (
	"os"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/server"
)

func main() {
	cfg, logger := config.New()

	if cfg.ADLService == "" {
		logger.Error("a service name for the Authorization Decision Log is required to start the PDP: set PDP_ADL_SERVICE_NAME or --adl-service")
		os.Exit(1)
	}

	if cfg.ADLDatabaseURL == "" {
		logger.Error("a Postgres URI for the Authorization Decision Log is required to start the PDP: set PDP_ADL_PG_URL or --adl-pg-url")
		os.Exit(1)
	}

	srv := server.NewService(cfg, logger)
	srv.Serve()
}
