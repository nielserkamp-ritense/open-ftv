package decisions

import (
	"encoding/base64"

	"github.com/goccy/go-json"
)

// BodyFromDecision builds the Logius ADL body object from a decision.
func BodyFromDecision(d *Decision) any {
	if d == nil || (d.Request == nil && d.Response == nil) {
		return nil
	}

	body := make(map[string]any, 2)
	if d.Request != nil {
		body["request"] = d.Request
	}
	if d.Response != nil {
		body["response"] = d.Response
	}
	return body
}

// RequestResponseFromBody extracts request and response from a stored body value.
func RequestResponseFromBody(body any) (request, response map[string]any) {
	m, ok := body.(map[string]any)
	if !ok {
		return nil, nil
	}

	return decodeBodyPart(m["request"]), decodeBodyPart(m["response"])
}

func decodeBodyPart(v any) map[string]any {
	switch x := v.(type) {
	case map[string]any:
		return x
	case string:
		return decodeBodyJSON([]byte(x))
	case []byte:
		return decodeBodyJSON(x)
	default:
		return nil
	}
}

func decodeBodyJSON(raw []byte) map[string]any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}

	data := raw
	if decoded, err := base64.StdEncoding.DecodeString(string(raw)); err == nil && json.Valid(decoded) {
		data = decoded
	}

	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	return out
}

// DefaultADLAttributes returns Logius attribute refs for body request/response.
func DefaultADLAttributes() map[string]any {
	return map[string]any{
		"adl.core.request":  map[string]any{"ref": "body.request"},
		"adl.core.response": map[string]any{"ref": "body.response"},
	}
}

func applyBodyAttribute(raw string, d *Decision) {
	var body struct {
		Request  json.RawMessage `json:"request"`
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		return
	}
	if len(body.Request) > 0 && string(body.Request) != "null" {
		d.Request = []byte(body.Request)
	}
	if len(body.Response) > 0 && string(body.Response) != "null" {
		d.Response = []byte(body.Response)
	}
}
