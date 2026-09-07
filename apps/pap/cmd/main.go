// Package main contains the main function for the app.
package main

// force google.golang.org/genproto v0.0.0-20250303144028-a0af3efb3deb to stay in go.mod
import (
	"os"

	_ "google.golang.org/genproto/protobuf/ptype"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pap/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pap/server"
)

func main() {
	cfg, logger := config.New()

	if err := cfg.Validate(); err != nil {
		logger.Error("invalid persistence configuration", "error", err)
		os.Exit(1)
	}

	server.NewService(cfg, logger).Serve()
}
