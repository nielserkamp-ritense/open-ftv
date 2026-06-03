package fiber

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server"
)

// Router is the function signature for setting up the service endpoints and custom middleware.
type Router func(ctx context.Context, svc *fiber.App)

// New initializes an HTTP(S) service (implemented with fiber and fasthttp).
func New(logger *slog.Logger, initRoutes Router, opts ...server.Option) server.Service {
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
		ErrorHandler:          ErrorHandler(logger),
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
	})

	s.defaultMiddleware()

	if initRoutes != nil {
		initRoutes(s.ctx, s.svc)
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

// Serve runs the HTTP(S) service.
func (s *service) Serve() {
	s.run()
	s.cancel()
}

// ServeWithWG runs the HTTP(S) service, and marks the WaitGroup done when finished.
func (s *service) ServeWithWG(wg *sync.WaitGroup) {
	defer wg.Done()
	s.run()
	s.cancel()
}

// Shutdown stops the HTTP(S) service.
func (s *service) Shutdown() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	select {
	case s.intChan <- syscall.SIGQUIT:
	case <-s.ctx.Done():
	}
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
