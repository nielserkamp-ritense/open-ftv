// Package adl implements authorization-decision logging conforming to the
// Authorization Decision Log (ADL) standard.
//
// The central type is Record, an implementation-independent, OpenTelemetry-shaped
// log record describing a single authorization decision. A Logger (see logger.go)
// implements the authlog.Logger interface: it maps each authlog.AuthRecord onto an
// ADL Record and persists exactly one record per evaluation - a durable write-ahead
// log (synchronous) followed by an asynchronous flush to one or more sinks.
package adl

// Status describes the outcome of the PDP's evaluation attempt.
//
// A successful evaluation that resulted in denial (decision:false, or an empty
// results array) MUST NOT be reported as Error: denial is a valid outcome and is
// therefore Ok (or Unset). Error is reserved for the PDP failing to evaluate.
type Status string

// The three conformant status values.
const (
	StatusUnset Status = "Unset" // default: evaluation completed without internal error.
	StatusOk    Status = "Ok"    // evaluation completed successfully (functionally equal to Unset).
	StatusError Status = "Error" // the PDP could not produce a decision.
)

// The five conformant event_name values, each mapping one-to-one onto an AuthZEN API.
const (
	EventAccessEvaluation  = "adl.access_evaluation"  // AuthZEN Access Evaluation API.
	EventAccessEvaluations = "adl.access_evaluations" // AuthZEN Access Evaluations API (batch).
	EventSearchSubject     = "adl.search_subject"     // AuthZEN Subject Search API.
	EventSearchAction      = "adl.search_action"      // AuthZEN Action Search API.
	EventSearchResource    = "adl.search_resource"    // AuthZEN Resource Search API.
)

// Attribute and body keys defined by the ADL standard.
const (
	KeyRequest       = "adl.core.request"
	KeyResponse      = "adl.core.response"
	KeyPolicies      = "adl.core.policies"
	KeyInformation   = "adl.core.information"
	KeyConfiguration = "adl.core.configuration"
	KeyTransactionID = "adl.fsc.transaction_id"

	// Data-subject reference keys (DPL / LDV linkage). These are source-style
	// references the application supplies on the request context; OpenFTV does not
	// implement the LDV itself, it only promotes the reference into the record.
	KeyDataSubjectID   = "dpl.core.data_subject_id"
	KeyDataSubjectType = "dpl.core.data_subject_type"

	// LabelLogRecordUID is the OpenTelemetry log.record.uid attribute. On the OTLP logs
	// path it carries the record's idempotency key so a backend can deduplicate
	// redelivered records (OTLP has no server-assigned document id of its own).
	LabelLogRecordUID = "log.record.uid"
)

// Record is an ADL authorization-decision log record.
//
// Field placement follows the standard's location rule: each adl.core.* field is
// carried in exactly one of Body (raw payload) or Attributes (source reference),
// never both.
type Record struct {
	// TraceID is the 32 lowercase-hex identifier of the trace this decision belongs to.
	TraceID string `json:"trace_id"`
	// SpanID is the 16 lowercase-hex identifier of the span representing this decision.
	SpanID string `json:"span_id"`
	// ParentSpanID is the 16 lowercase-hex parent span id; omitted only for a trace root.
	ParentSpanID string `json:"parent_span_id,omitempty"`
	// EventName identifies the type of authorization decision (one of the Event* values).
	EventName string `json:"event_name"`
	// Timestamp is the moment the decision was made, in milliseconds since the Unix epoch.
	Timestamp uint64 `json:"timestamp"`
	// Status is the outcome of the PDP's evaluation attempt.
	Status Status `json:"status"`
	// Attributes carries source references and metadata (adl.core.* references live here).
	Attributes map[string]any `json:"attributes,omitempty"`
	// Resource identifies the producer of the record.
	Resource map[string]any `json:"resource,omitempty"`
	// Body carries raw adl.core.* payloads (request, response, configuration, ...).
	Body map[string]any `json:"body,omitempty"`
}

// IdempotencyKey returns the (trace_id, span_id) idempotency key for the record.
func (r *Record) IdempotencyKey() string {
	return r.TraceID + ":" + r.SpanID
}
