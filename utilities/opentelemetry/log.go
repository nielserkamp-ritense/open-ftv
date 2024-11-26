package opentelemetry

import (
	"context"
	"fmt"
	"os"
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

// Logger represents the interface for logging events with OpenTelemetry.
//
// E.g. this can be used to log policy decisions into Logboek DataVerwerkingen.
type Logger interface {
	StartSpan(ctx context.Context, id string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
	Shutdown(ctx context.Context) error
}

// New instantiates a new logger for Logboek Dataverwerkingen.
//
// The given service name is mandatory.
// If the given url is empty, events will be logged to stdout.
// If batchTimeout is 0, it will be set to 5 seconds.
//
// The returned Logger does not use the global variables of OpenTelemetry.
// So it is safe to use multiple instances of Logger if the need arises.
// And it is safe to use this Logger in an app that has its own OpenTelemetry logic.
func New(service, url string, batchTimeout time.Duration, opts ...trace.TracerOption) (Logger, error) {
	if service == "" {
		return nil, fmt.Errorf("service name is required")
	}

	var exporter sdktrace.SpanExporter
	var err error

	if url == "" {
		exporter, _ = stdouttrace.New(
			stdouttrace.WithWriter(os.Stdout), // so we can silence it in unit-test :)
			stdouttrace.WithPrettyPrint(),
		)
	} else {
		if exporter, err = otlptracegrpc.New(
			context.TODO(),
			otlptracegrpc.WithEndpoint(url),
			otlptracegrpc.WithInsecure(),
		); err != nil {
			return nil, fmt.Errorf("failed to initialize trace exporter: %w", err)
		}
	}

	tr := resource.NewSchemaless(
		semconv.ServiceName(service),
		attribute.String("kind", "server"),
	)

	if batchTimeout == 0 {
		batchTimeout = 5 * time.Second
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(batchTimeout)),
		sdktrace.WithResource(tr),
	)

	return &logger{
		tp:     tp,
		tracer: tp.Tracer(service, opts...),
	}, nil
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
