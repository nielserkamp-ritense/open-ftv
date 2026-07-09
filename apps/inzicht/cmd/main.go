// Package main is the entrypoint for the Inzicht service.
package main

import (
	"context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/inzicht/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/inzicht/server"
)

func main() {
	cfg, logger := config.New()

	ctx := context.Background()
	svc, err := server.NewService(ctx, cfg, logger)
	if err != nil {
		logger.Error("failed to initialise inzicht service", "error", err)
		panic(err)
	}
	defer func() { _ = svc.Close() }()

	if err := svc.Serve(); err != nil {
		logger.Error("inzicht service stopped", "error", err)
		panic(err)
	}
}
