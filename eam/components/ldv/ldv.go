// Package ldv contains functionality regarding Logboek Dataverwerkingen.
package ldv

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// LDV represents the interface for logging messages to Logboek Dataverwerkingen.
type LDV interface {
	StartSpan(ctx context.Context, attributes ...attribute.KeyValue) (context.Context, trace.Span)
	Shutdown(ctx context.Context) error
}
