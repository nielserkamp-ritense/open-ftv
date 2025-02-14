// Package main contains the main function for the app.
package main

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pip/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pip/server"
)

func main() {
	cfg, logger := config.New()
	server.NewService(cfg, logger).Serve()
}
