package adl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Loki / OTel label keys and decision label values.
//
// These are the low-cardinality labels under which ADL records are indexed in an
// OpenTelemetry-backed store (Loki). They are deliberately kept small: high-cardinality
// fields such as trace_id, span_id or the subject id are NOT labels - they live in the
// record body (the log line) and are filtered client-side by the query layer. Turning
// them into labels would create an unbounded number of Loki streams.
const (
	LabelJob         = "job"          // constant stream selector ("adl"); always present.
	LabelEventName   = "event_name"   // adl event_name (5 fixed values -> bounded cardinality).
	LabelDecision    = "decision"     // permit / deny / unset (bounded cardinality).
	LabelServiceName = "service_name" // producing service (from resource service.name).

	JobValue = "adl" // value of the LabelJob stream selector.

	DecisionPermit = "permit"
	DecisionDeny   = "deny"
	DecisionUnset  = "unset"
)

// otlpLogSink flushes ADL records to an OpenTelemetry Collector over OTLP/HTTP as
// OpenTelemetry *log records* (the logs signal, not traces). This is the transport that
// lands records in Loki losslessly:
//
//   - the whole ADL Record is serialised as JSON and carried verbatim as the log record
//     Body (a single string value). The collector's Loki exporter writes the Body as the
//     Loki log line, so the query layer can JSON-decode it straight back into a Record -
//     a guaranteed lossless round-trip, identical to the write-ahead-log JSONL form.
//   - the low-cardinality label fields (event_name, decision, service_name) are mirrored
//     as log-record attributes so the collector can promote them to Loki labels.
//   - trace_id / span_id are set as the OTLP LogRecord TraceId / SpanId (and never as
//     labels), preserving trace continuity without exploding label cardinality.
//
// It uses OTLP/JSON, which the collector's otlphttp receiver accepts, so the sink needs
// no OpenTelemetry logs SDK dependency.
type otlpLogSink struct {
	endpoint string
	client   *http.Client
	scope    string
}

// OTLPLogConfig configures an OTLP/HTTP log sink.
type OTLPLogConfig struct {
	// Endpoint is the collector base URL (e.g. http://collector:4318) or a full
	// logs URL. When it has no path, "/v1/logs" is appended.
	Endpoint string
	// Client is the HTTP client used for export; defaults to a 15s-timeout client.
	Client *http.Client
	// Scope is the instrumentation scope name recorded on the logs; defaults to "adl".
	Scope string
}

// NewOTLPLogSink instantiates an ADL sink that exports records as OTLP/HTTP JSON logs.
func NewOTLPLogSink(cfg OTLPLogConfig) (Sink, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("adl: OTLP log sink requires an endpoint")
	}
	if err := requireSecureEndpoint(cfg.Endpoint); err != nil {
		return nil, err
	}
	client := cfg.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	scope := cfg.Scope
	if scope == "" {
		scope = "adl"
	}
	return &otlpLogSink{endpoint: logsURL(cfg.Endpoint), client: client, scope: scope}, nil
}

// logsURL appends the OTLP logs path when the endpoint carries none.
func logsURL(endpoint string) string {
	trimmed := strings.TrimRight(endpoint, "/")
	if strings.Contains(strings.TrimPrefix(strings.TrimPrefix(trimmed, "https://"), "http://"), "/") {
		return trimmed // already has a path.
	}
	return trimmed + "/v1/logs"
}

// Emit implements the Sink interface.
func (s *otlpLogSink) Emit(ctx context.Context, rec *Record) error {
	payload, err := EncodeOTLPLog(rec, s.scope)
	if err != nil {
		return fmt.Errorf("adl: encode OTLP log: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("adl: build OTLP log request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("adl: OTLP log export: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("adl: OTLP log export status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// --- OTLP/JSON encoding ------------------------------------------------------

// EncodeOTLPLog serialises a Record into an OTLP/HTTP JSON ExportLogsServiceRequest.
// The record's full JSON is the log Body; the label fields are attributes. It is exported
// so tests (and callers wanting the exact wire form) can build the same bytes the sink sends.
func EncodeOTLPLog(rec *Record, scope string) ([]byte, error) {
	line, err := json.Marshal(rec)
	if err != nil {
		return nil, err
	}

	attrs := []otlpKV{
		strKV(LabelEventName, rec.EventName),
		strKV(LabelDecision, DecisionLabel(rec)),
		// I1 on the OTLP logs path: the protocol assigns no server-side document id, so
		// deduplication of redelivered records is the backend's responsibility. We carry
		// the (trace_id:span_id) idempotency key as the standard log.record.uid attribute
		// so a backend can dedupe on it; the WAL/OpenSearch paths dedupe deterministically.
		strKV(LabelLogRecordUID, rec.IdempotencyKey()),
	}
	if sn := ServiceName(rec); sn != "" {
		attrs = append(attrs, strKV(LabelServiceName, sn))
	}
	if rec.Attributes != nil {
		if tx, ok := rec.Attributes[KeyTransactionID].(string); ok && tx != "" {
			attrs = append(attrs, strKV(KeyTransactionID, tx))
		}
	}

	lr := otlpLogRecord{
		TimeUnixNano:         strconv.FormatUint(rec.Timestamp*uint64(time.Millisecond), 10),
		ObservedTimeUnixNano: strconv.FormatUint(rec.Timestamp*uint64(time.Millisecond), 10),
		Body:                 otlpAnyValue{StringValue: ptr(string(line))},
		Attributes:           attrs,
		TraceID:              rec.TraceID,
		SpanID:               rec.SpanID,
	}

	req := otlpLogsPayload{ResourceLogs: []otlpResourceLogs{{
		Resource:  otlpResource{Attributes: resourceAttrs(rec.Resource)},
		ScopeLogs: []otlpScopeLogs{{Scope: otlpScope{Name: scope}, LogRecords: []otlpLogRecord{lr}}},
	}}}
	return json.Marshal(req)
}

// DecisionLabel returns the bounded-cardinality decision label for a record.
func DecisionLabel(rec *Record) string {
	if rec.Body != nil {
		if resp, ok := rec.Body[KeyResponse].(map[string]any); ok {
			if d, ok := resp["decision"].(bool); ok {
				if d {
					return DecisionPermit
				}
				return DecisionDeny
			}
		}
	}
	return DecisionUnset
}

// ServiceName returns the resource service.name, or "".
func ServiceName(rec *Record) string {
	if rec.Resource != nil {
		if s, ok := rec.Resource["service.name"].(string); ok {
			return s
		}
	}
	return ""
}

func resourceAttrs(res map[string]any) []otlpKV {
	if len(res) == 0 {
		return nil
	}
	out := make([]otlpKV, 0, len(res))
	for k, v := range res {
		if s, ok := v.(string); ok {
			out = append(out, strKV(k, s))
			continue
		}
		if raw, err := json.Marshal(v); err == nil {
			out = append(out, strKV(k, string(raw)))
		}
	}
	return out
}

type otlpLogsPayload struct {
	ResourceLogs []otlpResourceLogs `json:"resourceLogs"`
}

type otlpResourceLogs struct {
	Resource  otlpResource    `json:"resource"`
	ScopeLogs []otlpScopeLogs `json:"scopeLogs"`
}

type otlpResource struct {
	Attributes []otlpKV `json:"attributes,omitempty"`
}

type otlpScopeLogs struct {
	Scope      otlpScope       `json:"scope"`
	LogRecords []otlpLogRecord `json:"logRecords"`
}

type otlpScope struct {
	Name string `json:"name"`
}

type otlpLogRecord struct {
	TimeUnixNano         string       `json:"timeUnixNano"`
	ObservedTimeUnixNano string       `json:"observedTimeUnixNano"`
	Body                 otlpAnyValue `json:"body"`
	Attributes           []otlpKV     `json:"attributes,omitempty"`
	TraceID              string       `json:"traceId,omitempty"`
	SpanID               string       `json:"spanId,omitempty"`
}

type otlpKV struct {
	Key   string       `json:"key"`
	Value otlpAnyValue `json:"value"`
}

type otlpAnyValue struct {
	StringValue *string `json:"stringValue,omitempty"`
}

func strKV(k, v string) otlpKV { return otlpKV{Key: k, Value: otlpAnyValue{StringValue: ptr(v)}} }

func ptr[T any](v T) *T { return &v }

var _ Sink = (*otlpLogSink)(nil)
