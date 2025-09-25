// Package main contains the main function for a Policy Decision Point using AuthZEN.
package main

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/server"
)

func main() {
	cfg, logger := config.New()
	srv := server.NewService(cfg, logger)
	if !cfg.Migration.ExitAfter {
		srv.Serve()
	}
}
