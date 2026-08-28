// Package main contains the main function for the app.
package main

import (
	"os"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/server"
)

func main() {
	cfg, logger := config.New()

	if err := cfg.Validate(); err != nil {
		logger.Error("invalid persistence configuration", "error", err)
		os.Exit(1)
	}

	srv := server.NewExternal(cfg, logger)
	if !cfg.ExitAfter {
		srv.Serve()
	}
}
