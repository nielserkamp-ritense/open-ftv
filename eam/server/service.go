// Package server contains functionality for running a http/https service.
package server

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

// Service represents the interface for an HTTP service.
type Service interface {
	Context() context.Context
	Logger() *slog.Logger
	Serve() error // fails when the service could not start or stopped serving by itself.
	Shutdown()
	GetFiberApp() *fiber.App
}
