package opentelemetry

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
)

// Logger represents the interface for logging events to Logboek Dataverwerkingen using OpenTelemetry.
//
// E.g. this can be used to log policy decisions into Logboek DataVerwerkingen.
type Logger interface {
	StartSpan(ctx context.Context, id string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
	Shutdown(ctx context.Context) error
}

// LoggerConfig contains the configuration parameters for instantiating a new event logger for Logboek Dataverwerkingen.
type LoggerConfig struct {
	Service      string               // name of the service.
	URL          string               // URL of OpenTelemtry sink, or, "stdout", "stderr" or "slog".
	PrettyPrint  bool                 // enforces a pretty format when printing an event.
	Logger       *slog.Logger         // logger used when url == "slog".
	BatchTimeout time.Duration        // timeout for batching events.
	Opts         []trace.TracerOption // additional options for the tracer.
}

// New instantiates a new event logger for Logboek Dataverwerkingen.
//
// The given service name is mandatory.
//
// If the given url is empty, events will be printed on stdout.
// If the given url equals "stdout" or "stderr", events will be printed on the corresponding output.
// If the given url equals "log" or "slog", events will be formatted for and sent to the given logger.
// Any other value should be a valid URL pointing to an OpenTelemetry sink.
//
// If batchTimeout is 0, it will be set to 5 seconds.
//
// The returned Logger does not use the global variables of OpenTelemetry.
// So it is safe to use multiple instances of Logger if the need arises.
// And it is safe to use this Logger in an app that has its own OpenTelemetry logic.
func New(cfg *LoggerConfig) (Logger, error) {
	if cfg.Service == "" {
		return nil, fmt.Errorf("service name is required")
	}

	var exporter sdktrace.SpanExporter
	var err error

	switch strings.ToLower(cfg.URL) {
	case "", "stdout":
		exporter = newStdLogger(os.Stdout, cfg.PrettyPrint)

	case "stderr":
		exporter = newStdLogger(os.Stderr, cfg.PrettyPrint)

	case "log", "slog":
		if cfg.Logger == nil {
			return nil, fmt.Errorf("logger is required")
		}
		exporter = &slogLogger{logger: cfg.Logger}

	default:
		if exporter, err = otlptracegrpc.New(
			context.Background(),
			otlptracegrpc.WithEndpoint(cfg.URL),
			otlptracegrpc.WithInsecure(),
		); err != nil {
			return nil, fmt.Errorf("failed to initialize trace exporter: %w", err)
		}
	}

	tr := resource.NewSchemaless(
		semconv.ServiceName(cfg.Service),
		attribute.String("kind", "server"),
	)

	if cfg.BatchTimeout == 0 {
		cfg.BatchTimeout = 5 * time.Second
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(cfg.BatchTimeout)),
		sdktrace.WithResource(tr),
	)

	return &logger{
		tp:     tp,
		tracer: tp.Tracer(cfg.Service, cfg.Opts...),
	}, nil
}

func newStdLogger(f io.Writer, prettyPrint bool) *stdouttrace.Exporter {
	opts := []stdouttrace.Option{stdouttrace.WithWriter(f)}
	if prettyPrint {
		opts = append(opts, stdouttrace.WithPrettyPrint())
	}

	exporter, _ := stdouttrace.New(opts...)
	return exporter
}

// StartSpan initializes a new trace span.
func (l *logger) StartSpan(ctx context.Context, id string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return l.tracer.Start(ctx, id, opts...)
}

// Shutdown cleans up held resources.
//
// More specifically, it shuts down the OpenTelemetry trace provider.
//
// Note: StartSpan should never be called after this!
func (l *logger) Shutdown(ctx context.Context) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.tp == nil {
		return nil
	}

	err := l.tp.Shutdown(ctx)
	l.tp = nil
	l.tracer = nil
	return err
}

type logger struct {
	tp     *sdktrace.TracerProvider
	tracer trace.Tracer
	mutex  sync.Mutex
}
