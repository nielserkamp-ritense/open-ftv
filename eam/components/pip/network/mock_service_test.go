package network

import (
	"log/slog"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server"
	fiber2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server/fiber"
)

func newService(t *testing.T, logger *slog.Logger, ca, cert, key string, path string, h func(req *fiber.Ctx) error) server.Service {
	cfg := []server.ServerOption{
		server.WithDefaults(),
		server.WithAppName("mock service"),
		server.WithHostPort("127.0.0.1", 9000),
		server.WithRecovery(),
		server.WithSecurity(),
	}

	if cert != "" && key != "" {
		cfg = append(cfg, server.WithTLS(ca, cert, key))
		if ca != "" {
			cfg = append(cfg, server.WithMutualTLS())
		}
	}

	router := func(app *fiber.App) {
		v1 := app.Group("/v1")
		v1.Get(path, h)
	}

	s := fiber2.New(logger, router, cfg...)
	require.NotNil(t, s)

	return s
}
