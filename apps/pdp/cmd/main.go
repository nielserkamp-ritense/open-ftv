// Package main contains the main function for a Policy Decision Point using AuthZEN.
package main

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pdp/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pdp/server"
)

func main() {
	cfg, logger := config.New()
	server.NewService(cfg, logger).Serve()
}
