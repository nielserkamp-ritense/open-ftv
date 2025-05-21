package main

import (
	"os"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/generic/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/generic/server"
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
