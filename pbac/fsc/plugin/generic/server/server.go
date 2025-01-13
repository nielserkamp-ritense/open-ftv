// Package server handles the HTTP service component of the FSC Auth plugin.
package server

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components/ldv"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/handlers"
)

// Service represents the interface for an HTTP service.
type Service interface {
	Serve()
	Shutdown()
}

// NewService initializes an HTTP service (implemented with fiber/fasthttp).
func NewService(cfg *config.Config, logger *slog.Logger, logboek ldv.LDV) Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &service{ctx: ctx, cancel: cancel, cfg: cfg, logger: logger, logboek: logboek}
}

// Serve runs the HTTP service.
func (s *service) Serve() {
	s.mutex.Lock()
	s.intChan = make(chan os.Signal)
	signal.Notify(s.intChan, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	s.mutex.Unlock()

	s.svc = fiber.New(fiber.Config{
		CaseSensitive:         true,
		DisableDefaultDate:    true,
		DisableStartupMessage: true,
		BodyLimit:             s.cfg.MaxBody,
		ReadTimeout:           s.cfg.ReadTimeout,
		WriteTimeout:          s.cfg.WriteTimeout,
		IdleTimeout:           s.cfg.IdleTimeout,
		ErrorHandler:          s.errorHandler,
		AppName:               config.AppName,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		RequestMethods:        fiber.DefaultMethods,
	})

	s.initMiddleware()
	s.initRoutes()
	s.run()
	s.cancel()
}

// Shutdown can be used to stop the HTTP service.
func (s *service) Shutdown() {
	s.mutex.Lock()
	s.intChan <- syscall.SIGQUIT
	s.mutex.Unlock()
}

// errorHandler is the default error handler for things gone awry in fiber.
// E.g. invalid paths, bad parameters, code panics, etc.
// The actual error gets logged while the error response body is based on the embedded status code.
// If the error does not embed a status code, InternalServerError will be used as the status.
func (s *service) errorHandler(req *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError

	var e *fiber.Error
	if errors.As(err, &e) {
		status = e.Code
	}

	s.logger.Error("internal server error", "status", status, "error", err)
	return handlers.SendBasicResponse(req, status)
}

type service struct {
	ctx      context.Context
	cancel   context.CancelFunc
	cfg      *config.Config
	logger   *slog.Logger
	svc      *fiber.App
	logboek  ldv.LDV
	intChan  chan os.Signal
	shutdown atomic.Bool
	mutex    sync.Mutex
}
