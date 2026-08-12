package decisions

import "github.com/goccy/go-json"

// bodyRequestKey and bodyResponseKey are the literal Logius ADL body keys
// (Section 3.3.8): the raw adl.core.* payloads, as opposed to the source
// references an implementation may place under attributes at higher levels.
const (
	bodyRequestKey  = "adl.core.request"
	bodyResponseKey = "adl.core.response"
)

// jsonNull is the JSON encoding of a null value.
const jsonNull = "null"

// BodyFromDecision builds the Logius ADL body object from a decision.
func BodyFromDecision(d *Decision) any {
	if d == nil || (d.Request == nil && d.Response == nil) {
		return nil
	}

	body := make(map[string]any)
	if d.Request != nil {
		body[bodyRequestKey] = d.Request
	}

	if d.Response != nil {
		body[bodyResponseKey] = d.Response
	}

	return body
}

// RequestResponseFromBody extracts request and response from a stored body value.
func RequestResponseFromBody(body any) (request, response map[string]any) {
	m, ok := body.(map[string]any)
	if !ok {
		return nil, nil
	}

	request, _ = m[bodyRequestKey].(map[string]any)
	response, _ = m[bodyResponseKey].(map[string]any)

	return request, response
}

func applyBodyAttribute(raw string, d *Decision) {
	var body struct {
		Request  json.RawMessage `json:"adl.core.request"`
		Response json.RawMessage `json:"adl.core.response"`
	}
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		return
	}

	if len(body.Request) > 0 && string(body.Request) != jsonNull {
		d.Request = body.Request
	}

	if len(body.Response) > 0 && string(body.Response) != jsonNull {
		d.Response = body.Response
	}
}
