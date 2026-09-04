package opentelemetry

import (
	"fmt"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Option is the function signature for options when instantiating a new event logger.
type Option func(l *Exporter)

// WithExporter sets up the logger with the given exporter.
func WithExporter(e sdktrace.SpanExporter) Option {
	return func(l *Exporter) {
		if e == nil {
			l.err = fmt.Errorf("exporter must be supplied")
		} else {
			l.exporter = e
		}
	}
}

// WithBatchTimeout sets the timeout for batching events.
func WithBatchTimeout(timeout time.Duration) Option {
	return func(l *Exporter) {
		if timeout <= 0 {
			l.err = fmt.Errorf("batch timeout must be positive")
		} else {
			l.timeout = timeout
		}
	}
}
