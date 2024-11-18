// Package server handles the HTTP service component of the FSC Auth plugin.
package server

import (
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/handlers"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
)

type Service interface {
	Serve()
	Shutdown()
}

// NewService initializes a HTTP service which is implemented with fiber/fasthttp.
func NewService(cfg *config.Config, logger *slog.Logger) Service {
	return &service{cfg: cfg, logger: logger}
}

// Serve runs the HTTP service.
func (s *service) Serve() {
	s.intChan = make(chan os.Signal)
	signal.Notify(s.intChan, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

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
		RequestMethods:        []string{fiber.MethodGet, fiber.MethodHead, fiber.MethodPost, fiber.MethodOptions},
	})

	s.initMiddleware()
	s.initRoutes()
	s.run()
}

// Shutdown can be used to stop the HTTP service.
func (s *service) Shutdown() {
	s.intChan <- syscall.SIGQUIT
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
	cfg      *config.Config
	logger   *slog.Logger
	svc      *fiber.App
	intChan  chan os.Signal
	shutdown atomic.Bool
}
