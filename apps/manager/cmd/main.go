// Package main contains the main function for the app.
package main

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/server"
)

func main() {
	cfg, logger := config.New()

	srv := server.NewExternal(cfg, logger)
	if !cfg.ExitAfter {
		srv.Serve()
	}
}
