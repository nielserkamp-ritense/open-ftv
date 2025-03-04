// Package main contains the main function for an FSC Auth plugin.
package main

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/fsc/plugin/generic/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/fsc/plugin/generic/server"
)

func main() {
	cfg, logger := config.New()
	server.NewService(cfg, logger).Serve()
}
