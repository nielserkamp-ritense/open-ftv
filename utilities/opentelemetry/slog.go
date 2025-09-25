package opentelemetry

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// ExportSpans implements the SpanSyncer interface.
func (sl *slogLogger) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	for _, s := range spans {
		sl.logger.Info(sl.msg, "service", sl.service, "event", logEntry{
			Name:                 s.Name(),
			SpanContext:          s.SpanContext(),
			Parent:               s.Parent(),
			SpanKind:             s.SpanKind().String(),
			StartTime:            s.StartTime(),
			EndTime:              s.EndTime(),
			Attributes:           s.Attributes(),
			Links:                s.Links(),
			Events:               s.Events(),
			Status:               s.Status(),
			InstrumentationScope: s.InstrumentationScope(),
			Resource:             s.Resource(),
			DroppedAttributes:    s.DroppedAttributes(),
			DroppedLinks:         s.DroppedLinks(),
			DroppedEvents:        s.DroppedEvents(),
			ChildSpanCount:       s.ChildSpanCount(),
		})
	}
	return nil
}

// Shutdown implements the SpanSyncer interface.
func (sl *slogLogger) Shutdown(_ context.Context) error {
	return nil
}

type slogLogger struct {
	logger  *slog.Logger
	service string
	msg     string
}

type logEntry struct {
	Name                 string                `json:"name,omitempty"`
	SpanContext          trace.SpanContext     `json:"spanContext,omitempty"`
	Parent               trace.SpanContext     `json:"parent,omitempty"`
	SpanKind             string                `json:"spanKind,omitempty"`
	StartTime            time.Time             `json:"startTime,omitempty"`
	EndTime              time.Time             `json:"endTime,omitempty"`
	Attributes           []attribute.KeyValue  `json:"attributes,omitempty"`
	Links                []sdktrace.Link       `json:"links,omitempty"`
	Events               []sdktrace.Event      `json:"events,omitempty"`
	Status               sdktrace.Status       `json:"status,omitempty"`
	InstrumentationScope instrumentation.Scope `json:"instrumentationScope"`
	Resource             *resource.Resource    `json:"resource,omitempty"`
	DroppedAttributes    int                   `json:"droppedAttributes,omitempty"`
	DroppedLinks         int                   `json:"droppedLinks,omitempty"`
	DroppedEvents        int                   `json:"droppedEvents,omitempty"`
	ChildSpanCount       int                   `json:"childSpanCount,omitempty"`
}
