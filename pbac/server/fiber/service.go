package fiber

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

	handlers "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/server"
)

// Router is the function signature for setting up the service endpoints.
type Router func(svc *fiber.App)

// New initializes an HTTP service (implemented with fiber and fasthttp).
func New(logger *slog.Logger, initRoutes Router, opts ...server.ServerOption) server.Service {
	ctx, cancel := context.WithCancel(context.Background())

	s := &service{
		ctx:     ctx,
		cancel:  cancel,
		logger:  logger,
		intChan: make(chan os.Signal),
	}

	s.LoadOptions(opts...)

	signal.Notify(s.intChan, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	s.svc = fiber.New(fiber.Config{
		AppName:               s.AppName,
		CaseSensitive:         true,
		DisableDefaultDate:    true,
		DisableStartupMessage: true,
		RequestMethods:        fiber.DefaultMethods,
		BodyLimit:             s.MaxBody,
		ReadTimeout:           s.ReadTimeout,
		WriteTimeout:          s.WriteTimeout,
		IdleTimeout:           s.IdleTimeout,
		ErrorHandler:          s.errorHandler,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
	})

	s.defaultMiddleware()

	if initRoutes != nil {
		initRoutes(s.svc)
	}

	return s
}

// Context returns the server context.
func (s *service) Context() context.Context {
	return s.ctx
}

// Logger returns the logger.
func (s *service) Logger() *slog.Logger {
	return s.logger
}

// Serve runs the HTTP service.
func (s *service) Serve() {
	s.run()
	s.cancel()
}

// Shutdown stops the HTTP service.
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
	server.Config
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *slog.Logger
	svc      *fiber.App
	intChan  chan os.Signal
	mutex    sync.Mutex
	shutdown atomic.Bool
}
