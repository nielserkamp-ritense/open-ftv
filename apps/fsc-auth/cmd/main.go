// Package main contains the main function for an FSC Auth plugin.
package main

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/fsc-auth/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/fsc-auth/server"
)

func main() {
	cfg, logger := config.New()
	server.NewService(cfg, logger).Serve()
}
