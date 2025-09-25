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
	Timestamp   time.Time       // timestamp of the decision in UTC.
	RequestType AuthRequestType // type of request.
	Request     any             // request in AuthZEN format.
	Response    any             // response in AuthZEN format.
	Policies    uint64          // version number of the active bundle.
	Information any             // (optional) external data used for the decision.
	Engine      any             // (optional) details of the PDP.
	TraceID     string          // (optional) parent trace-id (hexadecimal format).
	SpanID      string          // (optional) parent span-id (hexadecimal format).
}

// MarshalJSON implements the json.Marshaler interface.
func (d *Decision) MarshalJSON() ([]byte, error) {
	return json.Marshal(decisionJSON{
		Timestamp:   d.Timestamp.Format(time.RFC3339),
		RequestType: d.RequestType.String(),
		Request:     d.Request,
		Response:    d.Response,
		Policies:    d.Policies,
		Information: d.Information,
		Engine:      d.Engine,
		TraceID:     d.TraceID,
		SpanID:      d.SpanID,
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Decision) UnmarshalJSON(b []byte) error {
	var j decisionJSON
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}

	d.Timestamp = convert.AnyToDateTime(j.Timestamp)
	d.RequestType = translate[j.RequestType]
	d.Request = j.Request
	d.Response = j.Response
	d.Policies = j.Policies
	d.Information = j.Information
	d.Engine = j.Engine
	d.TraceID = j.TraceID
	d.SpanID = j.SpanID

	return nil
}

type decisionJSON struct {
	Timestamp   string `json:"timestamp"`
	RequestType string `json:"request_type"`
	Request     any    `json:"request"`
	Response    any    `json:"response"`
	Policies    uint64 `json:"policies"`
	Information any    `json:"information,omitempty"`
	Engine      any    `json:"engine,omitempty"`
	TraceID     string `json:"trace_id,omitempty"`
	SpanID      string `json:"span_id,omitempty"`
}
