package pip

import (
	"context"
	"time"

	"github.com/goccy/go-json"
	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// SourceRef records where an attribute or entity value came from.
//
// It implements the ADL "Logged" source-reference pattern: instead of copying the
// upstream payload into the decision log, the PIP keeps a reference (span_id + WARC
// file + version) that an Authorization Decision Log can resolve, indexed on
// trace_id + span_id, to the exact request/response stored in the WARC log.
type SourceRef struct {
	Kind         string    `json:"kind"`                   // "attribute" or "entity".
	Key          string    `json:"key"`                    // attribute key or entity UID.
	TraceID      string    `json:"traceId,omitempty"`      // W3C trace id (32 hex).
	SpanID       string    `json:"spanId,omitempty"`       // W3C span id (16 hex) of the logged exchange.
	ParentSpanID string    `json:"parentSpanId,omitempty"` // optional parent span id.
	WARCFile     string    `json:"warcFile,omitempty"`     // WARC file that holds the exchange.
	Version      string    `json:"version,omitempty"`      // upstream data version (e.g. CloudEvents dataversion).
	Sequence     string    `json:"sequence,omitempty"`     // optional monotonic sequence tag.
	Time         time.Time `json:"time"`                   // when the value was recorded.
}

// SourceReferencer is implemented by a PIP that tracks source references for its values.
//
// The ADL side (built by a separate component) resolves these on level 3 ("+ information
// sources"). It can obtain the concrete implementation with a type assertion on pip.PIP.
type SourceReferencer interface {
	// RecordSourceRef stores (or replaces) the source reference for ref.Key.
	RecordSourceRef(ref SourceRef)
	// SourceRef returns the source reference for the given attribute key or entity UID.
	SourceRef(key string) (SourceRef, bool)
	// ListSourceRefs returns all known source references.
	ListSourceRefs() []SourceRef
}

// sourceRefStore is a small KV-backed store for SourceRef values.
type sourceRefStore struct {
	ctx      context.Context
	client   store.Store
	basePath string
}

func newSourceRefStore(ctx context.Context, client store.Store, basePath string) *sourceRefStore {
	return &sourceRefStore{ctx: ctx, client: client, basePath: convert.ForceSuffix(basePath, PathSeparator)}
}

func (s *sourceRefStore) put(ref SourceRef) {
	if s == nil || ref.Key == "" {
		return
	}
	if ref.Time.IsZero() {
		ref.Time = time.Now().UTC()
	}
	b, err := json.Marshal(ref)
	if err != nil {
		return
	}
	_ = s.client.Put(s.ctx, s.basePath+ref.Key, b, writeOptions)
}

func (s *sourceRefStore) get(key string) (SourceRef, bool) {
	if s == nil {
		return SourceRef{}, false
	}
	kv, err := s.client.Get(s.ctx, s.basePath+key, readOptions)
	if err != nil || kv == nil {
		return SourceRef{}, false
	}
	var ref SourceRef
	if err = json.Unmarshal(kv.Value, &ref); err != nil {
		return SourceRef{}, false
	}
	return ref, true
}

func (s *sourceRefStore) list() []SourceRef {
	if s == nil {
		return nil
	}
	list, err := s.client.List(s.ctx, s.basePath, readOptions)
	if err != nil {
		return nil
	}
	out := make([]SourceRef, 0, len(list))
	for _, kv := range list {
		var ref SourceRef
		if json.Unmarshal(kv.Value, &ref) == nil {
			out = append(out, ref)
		}
	}
	return out
}

// RecordSourceRef implements the SourceReferencer interface.
func (p *pip) RecordSourceRef(ref SourceRef) {
	p.sourceRefs.put(ref)
	if p.logger != nil {
		p.logger.Debug("recorded source reference", "kind", ref.Kind, "key", ref.Key, "span", ref.SpanID, "warc", ref.WARCFile, "version", ref.Version)
	}
}

// SourceRef implements the SourceReferencer interface.
func (p *pip) SourceRef(key string) (SourceRef, bool) {
	return p.sourceRefs.get(key)
}

// ListSourceRefs implements the SourceReferencer interface.
func (p *pip) ListSourceRefs() []SourceRef {
	return p.sourceRefs.list()
}
