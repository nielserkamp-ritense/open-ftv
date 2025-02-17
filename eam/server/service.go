// Package server contains functionality for running a http/https service.
package server

import (
	"context"
	"log/slog"
)

// Service represents the interface for an HTTP service.
type Service interface {
	Context() context.Context // the context of the service.
	Logger() *slog.Logger     // the logger for the service.
	Serve()                   // start the service.
	Shutdown()                // stop the service.
}
