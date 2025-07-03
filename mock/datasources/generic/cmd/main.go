package main

import (
	"os"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/generic/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/generic/server"
)

func main() {
	cfg, logger, err := config.New()
	if err != nil {
		if logger != nil {
			logger.Error("failed to load configuration", "error", err)
			os.Exit(1)
		}
		panic(err)
	}

	logger.Info("configuration loaded successfully", "config", cfg)

	srv, err2 := server.NewService(cfg, logger)
	if err2 != nil {
		logger.Error("failed to start service", "error", err2)
		os.Exit(1)
	}

	srv.Serve()
}
