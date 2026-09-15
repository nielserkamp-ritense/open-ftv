package server

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/filesystem"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
)

const (
	indexFile      = "index.html"
	runtimeEnvPath = "/runtime-env.js"
)

func (s *Services) initMainRoutes(_ context.Context, svc *fiber.App) {
	svc.Get(runtimeEnvPath, s.getRuntimeEnv)

	// embedded files have no modification time, so without an ETag browsers download every asset again on each load.
	svc.Use(etag.New())

	// unknown paths get index.html, so reloading a client-side route still loads the app.
	svc.Use(filesystem.New(filesystem.Config{
		Root:         http.FS(s.assets),
		Index:        indexFile,
		NotFoundFile: indexFile,
	}))
}

func (s *Services) initHealthRoutes(_ context.Context, svc *fiber.App) {
	// liveness and readiness.
	svc.Get(handle.PathHealthZ, s.chk.HealthZ).
		Get(handle.PathLiveZ, s.chk.LiveZ).
		Get(handle.PathReadyZ, s.chk.ReadyZ)
}

func (s *Services) getRuntimeEnv(fc *fiber.Ctx) error {
	fc.Set(fiber.HeaderContentType, "application/javascript; charset=utf-8")
	fc.Set(fiber.HeaderCacheControl, "no-store")

	return fc.Send(s.runtimeEnv)
}
