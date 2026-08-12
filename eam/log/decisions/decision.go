package decisions

import (
	"time"

	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// Decision represents an entry for the Authorization Decision Log.
//
// See the current revision of the standard at:
// https://vng-realisatie.github.io/authorization-decision-log/
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
	TraceID      string          // W3C trace-id (hexadecimal format).
	SpanID       string          // W3C span-id for this log record (hexadecimal format).
	ParentSpanID string          // W3C parent span-id when this record is a child span.
}

// MarshalJSON implements the json.Marshaler interface.
func (d *Decision) MarshalJSON() ([]byte, error) {
	j := decisionJSON{
		Timestamp:    d.Timestamp.Format(time.RFC3339),
		RequestType:  d.RequestType.String(),
		EventName:    d.EventName,
		Status:       string(d.Status),
		Request:      d.Request,
		Response:     d.Response,
		Policies:     d.Policies,
		Information:  d.Information,
		Engine:       d.Engine,
		TraceID:      d.TraceID,
		SpanID:       d.SpanID,
		ParentSpanID: d.ParentSpanID,
	}
	if !d.Timestamp.IsZero() {
		j.TimestampMs = uint64(d.Timestamp.UnixMilli())
	}
	return json.Marshal(j)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Decision) UnmarshalJSON(b []byte) error {
	var j decisionJSON
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}

	d.Timestamp = convert.AnyToDateTime(j.Timestamp)
	d.RequestType = translate[j.RequestType]
	d.EventName = j.EventName
	d.Status = Status(j.Status)
	d.Request = j.Request
	d.Response = j.Response
	d.Policies = j.Policies
	d.Information = j.Information
	d.Engine = j.Engine
	d.TraceID = j.TraceID
	d.SpanID = j.SpanID
	d.ParentSpanID = j.ParentSpanID

	return nil
}

type decisionJSON struct {
	Timestamp    string `json:"timestamp"`
	TimestampMs  uint64 `json:"timestamp_ms,omitempty"`
	RequestType  string `json:"request_type"`
	EventName    string `json:"event_name,omitempty"`
	Status       string `json:"status,omitempty"`
	Request      any    `json:"request"`
	Response     any    `json:"response"`
	Policies     uint64 `json:"policies"`
	Information  any    `json:"information,omitempty"`
	Engine       any    `json:"engine,omitempty"`
	TraceID      string `json:"trace_id,omitempty"`
	SpanID       string `json:"span_id,omitempty"`
	ParentSpanID string `json:"parent_span_id,omitempty"`
}
