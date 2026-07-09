// Package cloudevents implements a minimal, dependency-free decoder for
// CloudEvents 1.0 (https://github.com/cloudevents/spec), as used by the PIP
// push-ingest endpoint.
//
// Both content modes of the HTTP protocol binding are supported:
//   - structured mode: Content-Type application/cloudevents+json, the full event as JSON body.
//   - binary mode: ce-* headers carry the context attributes, the body is the event data.
//
// The version tag of the carried data can be supplied via the extension attributes
// "dataversion" (free-form version label) and/or "sequence" (monotonic counter),
// retrievable via Event.DataVersion / Event.Sequence.
//
// Scope: this is plain CloudEvents 1.0, NOT an "NLGov profile" of CloudEvents. It
// validates only the four required core attributes (id, source, type, specversion)
// and passes extensions through untouched.
//
// TODO(nlgov): a genuine NLGov CloudEvents profile would additionally mandate/verify,
// among others: a governed vocabulary and URI form for `type` and `source`; required
// `subject`, `time` (RFC3339) and `datacontenttype`; a `dataschema` bound to a
// published schema (and validation of `data` against it); conventions/registration
// for the extension attributes used (e.g. `dataversion`, `sequence`, correlation/trace
// ids); and signing/provenance requirements. None of that is enforced here.
package cloudevents

import (
	"fmt"
	"strings"

	"github.com/goccy/go-json"
)

// Content types for the CloudEvents HTTP protocol binding.
const (
	ContentTypeStructured = "application/cloudevents+json"
	headerPrefix          = "ce-"
)

// Extension attribute names for data versioning.
const (
	ExtDataVersion = "dataversion"
	ExtSequence    = "sequence"
)

// Event represents a single CloudEvents 1.0 event.
type Event struct {
	SpecVersion     string            // required "specversion" (must be 1.0).
	ID              string            // required "id".
	Source          string            // required "source".
	Type            string            // required "type".
	Subject         string            // optional "subject".
	Time            string            // optional "time" (RFC3339).
	DataContentType string            // optional "datacontenttype".
	Extensions      map[string]string // any extension attributes (lowercase names).
	Data            []byte            // the event payload (raw bytes; JSON for structured data).
}

// DataVersion returns the "dataversion" extension attribute, or "".
func (e *Event) DataVersion() string { return e.Extensions[ExtDataVersion] }

// Sequence returns the "sequence" extension attribute, or "".
func (e *Event) Sequence() string { return e.Extensions[ExtSequence] }

// DedupKey returns the idempotency key for the event: (source, id).
func (e *Event) DedupKey() string { return e.Source + "\x00" + e.ID }

// Validate checks the required context attributes (id, source, type, specversion).
func (e *Event) Validate() error {
	switch {
	case e == nil:
		return fmt.Errorf("cloudevents: no event")
	case e.SpecVersion == "":
		return fmt.Errorf("cloudevents: missing required attribute 'specversion'")
	case e.SpecVersion != "1.0":
		return fmt.Errorf("cloudevents: unsupported specversion %q", e.SpecVersion)
	case e.ID == "":
		return fmt.Errorf("cloudevents: missing required attribute 'id'")
	case e.Source == "":
		return fmt.Errorf("cloudevents: missing required attribute 'source'")
	case e.Type == "":
		return fmt.Errorf("cloudevents: missing required attribute 'type'")
	}
	return nil
}

// IsStructured reports whether the given Content-Type indicates structured mode.
func IsStructured(contentType string) bool {
	ct, _, _ := strings.Cut(contentType, ";")
	return strings.EqualFold(strings.TrimSpace(ct), ContentTypeStructured)
}

// Parse decodes an event from an HTTP request, choosing structured or binary mode
// based on the Content-Type header. headers must contain all request headers with
// their original names; lookup is case-insensitive.
func Parse(contentType string, headers map[string]string, body []byte) (*Event, error) {
	if IsStructured(contentType) {
		return ParseStructured(body)
	}
	return ParseBinary(contentType, headers, body)
}

// ParseStructured decodes a structured-mode (application/cloudevents+json) event.
func ParseStructured(body []byte) (*Event, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("cloudevents: invalid JSON envelope: %w", err)
	}

	e := &Event{Extensions: map[string]string{}}

	for key, val := range raw {
		switch strings.ToLower(key) {
		case "specversion":
			e.SpecVersion = jsonString(val)
		case "id":
			e.ID = jsonString(val)
		case "source":
			e.Source = jsonString(val)
		case "type":
			e.Type = jsonString(val)
		case "subject":
			e.Subject = jsonString(val)
		case "time":
			e.Time = jsonString(val)
		case "datacontenttype":
			e.DataContentType = jsonString(val)
		case "data":
			e.Data = []byte(val)
		case "data_base64":
			var s string
			if json.Unmarshal(val, &s) == nil {
				e.Data = decodeBase64(s)
			}
		default:
			e.Extensions[strings.ToLower(key)] = jsonString(val)
		}
	}

	if e.DataContentType == "" && e.Data != nil {
		e.DataContentType = "application/json"
	}

	return e, e.Validate()
}

// ParseBinary decodes a binary-mode event: context attributes from ce-* headers,
// event data from the request body.
func ParseBinary(contentType string, headers map[string]string, body []byte) (*Event, error) {
	e := &Event{Extensions: map[string]string{}, DataContentType: contentType, Data: body}

	for key, val := range headers {
		lower := strings.ToLower(key)
		if !strings.HasPrefix(lower, headerPrefix) {
			continue
		}

		switch name := strings.TrimPrefix(lower, headerPrefix); name {
		case "specversion":
			e.SpecVersion = val
		case "id":
			e.ID = val
		case "source":
			e.Source = val
		case "type":
			e.Type = val
		case "subject":
			e.Subject = val
		case "time":
			e.Time = val
		default:
			e.Extensions[name] = val
		}
	}

	return e, e.Validate()
}

// jsonString renders a raw JSON value as its string form (unquoting JSON strings).
func jsonString(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return strings.TrimSpace(string(raw))
}
