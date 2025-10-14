// Package server contains functionality for running a http/https service.
package server

import (
	"context"
	"log/slog"
	"sync"
)

// Service represents the interface for an HTTP service.
type Service interface {
	Context() context.Context       // the context of the service.
	Logger() *slog.Logger           // the logger for the service.
	Serve()                         // start the service.
	ServeWithWG(wg *sync.WaitGroup) // start the service, and signal the WaitGroup when finished.
	Shutdown()                      // stop the service.
}
