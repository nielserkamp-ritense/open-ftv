// Package main contains the main function for the app.
package main

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pip/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pip/server"
)

func main() {
	cfg, logger := config.New()
	server.NewServices(cfg, logger).Serve()
}
