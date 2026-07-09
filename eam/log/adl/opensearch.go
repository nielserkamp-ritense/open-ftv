package adl

import (
	"context"
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities-no-ci/opensearch"
)

// openSearchSink flushes ADL records to an OpenSearch index, reusing the existing
// OpenSearch logging client. The ADL Record is stored verbatim as the document, so the
// index carries records in their standard shape.
type openSearchSink struct {
	index string
	sink  opensearch.Logger
}

// newOpenSearchLogger is overridable in unit tests.
var newOpenSearchLogger = opensearch.NewLogger

// NewOpenSearchSink instantiates an ADL sink backed by an OpenSearch cluster.
func NewOpenSearchSink(index, user, pswd string, endpoints ...string) (Sink, error) {
	for _, ep := range endpoints {
		if err := requireSecureEndpoint(ep); err != nil {
			return nil, err
		}
	}
	sink, err := newOpenSearchLogger(user, pswd, endpoints)
	if err != nil {
		return nil, fmt.Errorf("adl: failed to create OpenSearch sink: %w", err)
	}
	return &openSearchSink{index: index, sink: sink}, nil
}

// Emit implements the Sink interface.
//
// I1 (idempotent ingestion): the document ID is the record's (trace_id:span_id)
// idempotency key rather than a random UUID, so redelivery - a WAL replay or a retried
// flush, even after a process restart - overwrites the same document instead of creating
// a duplicate. The OpenSearch client indexes by this ID (an upsert), making ingestion
// idempotent across the whole cluster lifetime, not just within one process.
func (s *openSearchSink) Emit(ctx context.Context, rec *Record) error {
	return s.sink.Log(ctx, false, opensearch.LogRecord{Index: s.index, ID: rec.IdempotencyKey(), Data: rec})
}
