package decisions

import (
	"time"
)

// Decision represents an entry for the Authorization Decision Log.
//
// See the current revision of the standard at:
// https://logius-standaarden.github.io/authorization-decision-log/
type Decision struct {
	Timestamp    time.Time       // timestamp of the decision in UTC.
	RequestType  AuthRequestType // type of request.
	EventName    string          // Logius ADL event_name.
	Status       Status          // Logius ADL status.
	Request      any             // request in AuthZEN format.
	Response     any             // response in AuthZEN format.
	Policies     uint64          // version number of the active bundle.
	Information  any             // (optional) external data used for the decision.
	Engine       any             // (optional) details of the PDP.
	Resource     any             // (optional) Logius ADL resource: identifies the producer of the log record.
	TraceID      string          // W3C trace-id (hexadecimal format).
	SpanID       string          // W3C span-id for this log record (hexadecimal format).
	ParentSpanID string          // W3C parent span-id when this record is a child span.
}
