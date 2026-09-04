// Package main contains the main function for the app.
package main

import (
	"os"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pip/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pip/server"
)

func main() {
	cfg, logger := config.New()

	if cfg.DatabaseURL == "" {
		logger.Error("a Postgres URI is required to start the PIP: set PIP_DATABASE_URL or --database-url")
		os.Exit(1)
	}

	srv := server.NewServices(cfg, logger)
	srv.Serve()
}
