// Package opentelemetry contains functionality for working with OpenTelemetry modules.
package opentelemetry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
)

// Logger represents the interface for logging events using OpenTelemetry.
//
// This can be used to log policy decisions to the Authorization Decision Log.
type Logger interface {
	StartSpan(ctx context.Context, id string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
	Shutdown(ctx context.Context) error
}

// New instantiates a new event logger using OpenTelemetry.
//
// The given service name is mandatory.
//
// Use at most one of WithOT, WithFile, WithSLog or WithExporter to set up the exporter to use.
// Failing to do so will result in an error.
//
// The returned Logger does not use the global variables of OpenTelemetry.
// So it is safe to use multiple instances of Logger if the need arises.
// And it is safe to use this Logger in an app that has its own OpenTelemetry logic.
func New(ctx context.Context, service string, opts ...Option) (*Exporter, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if service == "" {
		return nil, fmt.Errorf("service name is required")
	}

	e := &Exporter{ctx: ctx, service: service, timeout: 5 * time.Second}
	for i := range opts {
		opts[i](e)
	}

	if e.err != nil {
		return nil, e.err
	}
	if e.exporter == nil {
		return nil, fmt.Errorf("no exporter configured for service %s", service)
	}

	tr := resource.NewSchemaless(
		semconv.ServiceName(e.service),
		attribute.String("kind", "server"),
	)

	e.tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(e.exporter, sdktrace.WithBatchTimeout(e.timeout)),
		sdktrace.WithResource(tr),
	)
	e.tracer = e.tp.Tracer(e.service, e.opts...)

	return e, nil
}

// StartSpan initializes a new trace span.
func (e *Exporter) StartSpan(ctx context.Context, id string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return e.tracer.Start(ctx, id, opts...)
}

// Shutdown cleans up held resources.
//
// More specifically, it shuts down the OpenTelemetry trace provider.
//
// Note: StartSpan should never be called after this!
func (e *Exporter) Shutdown(ctx context.Context) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.tp == nil {
		return nil
	}

	err := e.tp.Shutdown(ctx)
	e.tp = nil
	e.tracer = nil
	return err
}

// Exporter implements an event logger with OpenTelemetry.
type Exporter struct {
	ctx      context.Context
	service  string
	exporter sdktrace.SpanExporter
	err      error
	timeout  time.Duration
	opts     []trace.TracerOption
	tp       *sdktrace.TracerProvider
	tracer   trace.Tracer
	mutex    sync.Mutex
}
