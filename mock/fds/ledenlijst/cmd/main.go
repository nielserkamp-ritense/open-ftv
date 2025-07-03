package main

import (
	"os"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/fds/ledenlijst/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/fds/ledenlijst/server"
)

func main() {
	cfg, logger, err := config.New()
	if err != nil {
		if logger != nil {
			logger.Error("init failed", "error", err)
			os.Exit(1)
		}
		panic(err)
	}

	logger.Info("configuration loaded successfully", "config", cfg)
	server.NewService(cfg, logger).Serve()
}
