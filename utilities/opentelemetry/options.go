package opentelemetry

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Option is the function signature for options when instantiating a new event logger.
type Option func(l *Exporter)

// WithOT sets up the logger to use the given OpenTelemetry sink.
//
// This option is mutually exclusive with WithFile, WithSLog and WithExporter.
func WithOT(url string, insecure bool) Option {
	return func(l *Exporter) {
		if url == "" {
			l.err = fmt.Errorf("url must be filled")
		} else {
			opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(url)}
			if insecure {
				opts = append(opts, otlptracegrpc.WithInsecure())
			}

			l.exporter, l.err = otlptracegrpc.New(l.ctx, opts...)
		}
	}
}

// WithFile sets up the logger to write to a file; e.g., stdout or stderr.
//
// Set pretty to true for more detailed output.
//
// This option is mutually exclusive with WithOT, WithSLog and WithExporter.
func WithFile(f *os.File, pretty bool) Option {
	return func(l *Exporter) {
		if f == nil {
			l.err = fmt.Errorf("file must be valid")
		} else {
			opts := []stdouttrace.Option{stdouttrace.WithWriter(f)}
			if pretty {
				opts = append(opts, stdouttrace.WithPrettyPrint())
			}
			l.exporter, l.err = stdouttrace.New(opts...)
		}
	}
}

// WithSLog sets up the logger to write to the given logger.
//
// This option is mutually exclusive with WithOT, WithFile and WithFile.
func WithSLog(log *slog.Logger, msg string) Option {
	return func(l *Exporter) {
		if log == nil {
			l.err = fmt.Errorf("logger must be supplied")
		} else {
			l.exporter = &slogLogger{logger: log, service: l.service, msg: msg}
		}
	}
}

// WithExporter sets up the logger with the given exporter.
//
// This option is mutually exclusive with WithOT, WithFile and WithSLog.
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

// WithTraceOpts adds additional options when using an OpenTelemetry sink.
func WithTraceOpts(opts ...trace.TracerOption) Option {
	return func(l *Exporter) {
		l.opts = opts
	}
}
