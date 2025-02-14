// Package main contains the main function for the app.
package main

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/manager/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/manager/server"
)

func main() {
	cfg, logger := config.New()
	server.NewService(cfg, logger).Serve()
}
