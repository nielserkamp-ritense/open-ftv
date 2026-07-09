package adl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// richRecord builds a fully-populated Level-4-ish record exercising every field a Source
// must be able to reconstruct: body request/response including obligations, attribute
// source references, and a producer resource.
func richRecord() *Record {
	return &Record{
		TraceID:      "0af7651916cd43dd8448eb211c80319c",
		SpanID:       "b7ad6b7169203331",
		ParentSpanID: "00f067aa0ba902b7",
		EventName:    EventAccessEvaluation,
		Timestamp:    1751371200000, // 2025-07-01T12:00:00Z (ms).
		Status:       StatusOk,
		Attributes: map[string]any{
			KeyTransactionID: "tx-42",
			KeyPolicies:      map[string]any{"brp-2026": "v1.1.0"},
			KeyInformation:   []any{map[string]any{"source": "ldv", "span_id": "abcd"}},
		},
		Resource: map[string]any{"service.name": "pdp", "env": "prod"},
		Body: map[string]any{
			KeyRequest: map[string]any{
				"subject":  map[string]any{"type": "organisation", "id": "gemeente-amsterdam"},
				"action":   map[string]any{"name": "read"},
				"resource": map[string]any{"type": "brp", "id": "persoon/123"},
				"context":  map[string]any{"purpose": "huisvestingswet"},
			},
			KeyResponse: map[string]any{
				"decision": true,
				"context": map[string]any{
					"obligations": []any{
						map[string]any{"id": "log_access", "attributes": map[string]any{"level": float64(2)}},
					},
				},
			},
		},
	}
}

// normalize renders a value through a decode/encode cycle so numeric representations
// match regardless of int vs float64 origin.
func normalize(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var anyv any
	if err := json.Unmarshal(raw, &anyv); err != nil {
		t.Fatal(err)
	}
	out, err := json.Marshal(anyv)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

// decodeOTLPLine extracts the log-line string and label attributes from an OTLP/JSON
// ExportLogsServiceRequest - i.e. it does what the collector's Loki exporter does.
func decodeOTLPLine(t *testing.T, payload []byte) (string, map[string]string) {
	t.Helper()
	var req struct {
		ResourceLogs []struct {
			ScopeLogs []struct {
				LogRecords []struct {
					Body struct {
						StringValue string `json:"stringValue"`
					} `json:"body"`
					Attributes []struct {
						Key   string `json:"key"`
						Value struct {
							StringValue string `json:"stringValue"`
						} `json:"value"`
					} `json:"attributes"`
					TraceID string `json:"traceId"`
					SpanID  string `json:"spanId"`
				} `json:"logRecords"`
			} `json:"scopeLogs"`
		} `json:"resourceLogs"`
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		t.Fatalf("decode OTLP payload: %v", err)
	}
	if len(req.ResourceLogs) != 1 || len(req.ResourceLogs[0].ScopeLogs) != 1 ||
		len(req.ResourceLogs[0].ScopeLogs[0].LogRecords) != 1 {
		t.Fatalf("unexpected OTLP shape: %s", payload)
	}
	lr := req.ResourceLogs[0].ScopeLogs[0].LogRecords[0]
	labels := map[string]string{}
	for _, a := range lr.Attributes {
		labels[a.Key] = a.Value.StringValue
	}
	if lr.TraceID == "" || lr.SpanID == "" {
		t.Errorf("OTLP log record missing trace/span id")
	}
	return lr.Body.StringValue, labels
}

// TestEncodeOTLPLogRoundTrip proves the sink's export form is lossless: a Record encoded
// to the OTLP/JSON logs form and decoded back through the collector's Loki-exporter
// transform (log Body -> Loki line -> Record) reconstructs the original record exactly.
func TestEncodeOTLPLogRoundTrip(t *testing.T) {
	rec := richRecord()

	payload, err := EncodeOTLPLog(rec, "adl")
	if err != nil {
		t.Fatal(err)
	}

	line, labels := decodeOTLPLine(t, payload)

	// The line is the exact record JSON: decode it straight back into a Record.
	var got Record
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatalf("decode line: %v", err)
	}
	if normalize(t, &got) != normalize(t, rec) {
		t.Fatalf("round-trip mismatch:\n got:  %s\n want: %s", normalize(t, &got), normalize(t, rec))
	}

	// Labels are the chosen low-cardinality set, and carry no high-cardinality field.
	if labels[LabelEventName] != EventAccessEvaluation {
		t.Errorf("event_name label = %q", labels[LabelEventName])
	}
	if labels[LabelDecision] != DecisionPermit {
		t.Errorf("decision label = %q", labels[LabelDecision])
	}
	if labels[LabelServiceName] != "pdp" {
		t.Errorf("service_name label = %q", labels[LabelServiceName])
	}
	for _, forbidden := range []string{"trace_id", "span_id", "subject"} {
		if _, ok := labels[forbidden]; ok {
			t.Errorf("high-cardinality field %q must not be a label", forbidden)
		}
	}
}

// TestOTLPLogSinkEmit verifies the sink POSTs OTLP/JSON to /v1/logs and the payload
// round-trips back to the original record.
func TestOTLPLogSinkEmit(t *testing.T) {
	rec := richRecord()

	var captured []byte
	var gotPath, gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotCT = r.Header.Get("Content-Type")
		captured, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sink, err := NewOTLPLogSink(OTLPLogConfig{Endpoint: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if err := sink.Emit(context.Background(), rec); err != nil {
		t.Fatal(err)
	}

	if gotPath != "/v1/logs" {
		t.Errorf("path = %q, want /v1/logs", gotPath)
	}
	if gotCT != "application/json" {
		t.Errorf("content-type = %q", gotCT)
	}

	line, _ := decodeOTLPLine(t, captured)
	var got Record
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatal(err)
	}
	if normalize(t, &got) != normalize(t, rec) {
		t.Fatalf("sink round-trip mismatch")
	}
}

func TestDecisionLabel(t *testing.T) {
	permit := richRecord()
	if DecisionLabel(permit) != DecisionPermit {
		t.Errorf("permit label = %q", DecisionLabel(permit))
	}
	permit.Body[KeyResponse] = map[string]any{"decision": false}
	if DecisionLabel(permit) != DecisionDeny {
		t.Errorf("deny label = %q", DecisionLabel(permit))
	}
	permit.Body = nil
	if DecisionLabel(permit) != DecisionUnset {
		t.Errorf("unset label = %q", DecisionLabel(permit))
	}
}
