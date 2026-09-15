package main

import (
	"io/fs"
	"os"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/management-interface/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/management-interface/server"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/management-interface/web"
)

func main() {
	cfg, logger := config.New()

	assets, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		logger.Error("failed to open the embedded static assets", "error", err)
		os.Exit(1)
	}

	srv, err := server.NewService(cfg, logger, assets, config.RuntimeEnv())
	if err != nil {
		logger.Error("failed to start the management interface", "error", err)
		os.Exit(1)
	}

	srv.Serve()
}
