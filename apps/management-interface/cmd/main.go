package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/management-interface/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/management-interface/server"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/management-interface/web"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func main() {
	cfg, err := config.Load(os.Args[1:], os.LookupEnv)
	if errors.Is(err, flag.ErrHelp) {
		return
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load configuration:", err)
		os.Exit(1)
	}

	logger, err := slog2.Init(cfg.Log.Output, cfg.Log.Format, cfg.Log.Level, cfg.Log.Source)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to initialize logger:", err)
		os.Exit(1)
	}

	logger.Info("git version", "hash", config.GitHash, "tag", config.GitTag)

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
