// Package main contains the main function for the app.
package main

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/server"
)

func main() {
	cfg, logger := config.New()
	srv := server.NewService(cfg, logger)
	if !cfg.Migration.ExitAfter {
		srv.Serve()
	}
}
