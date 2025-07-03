// Package main contains the main function for the app.
package main

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pap/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pap/server"
)

func main() {
	cfg, logger := config.New()
	server.NewService(cfg, logger).Serve()
}
