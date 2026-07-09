package cloudevents

import (
	"testing"
)

func TestParseStructured(t *testing.T) {
	t.Parallel()

	body := []byte(`{
		"specversion": "1.0",
		"id": "evt-1",
		"source": "https://brp.example/personen",
		"type": "nl.ftv.pip.attributes.updated",
		"subject": "person/123",
		"time": "2026-07-08T10:00:00Z",
		"dataversion": "2.3.1",
		"sequence": "17",
		"datacontenttype": "application/json",
		"data": {"attributes":[{"key":"a","value":1}]}
	}`)

	e, err := ParseStructured(body)
	if err != nil {
		t.Fatalf("ParseStructured: %v", err)
	}

	if e.ID != "evt-1" || e.Source != "https://brp.example/personen" || e.Type != "nl.ftv.pip.attributes.updated" {
		t.Errorf("unexpected context attributes: %+v", e)
	}
	if e.DataVersion() != "2.3.1" {
		t.Errorf("DataVersion = %q, want 2.3.1", e.DataVersion())
	}
	if e.Sequence() != "17" {
		t.Errorf("Sequence = %q, want 17", e.Sequence())
	}
	if len(e.Data) == 0 {
		t.Error("expected data payload")
	}
}

func TestParseStructuredMissingRequired(t *testing.T) {
	t.Parallel()

	for name, body := range map[string]string{
		"missing id":          `{"specversion":"1.0","source":"s","type":"t"}`,
		"missing source":      `{"specversion":"1.0","id":"i","type":"t"}`,
		"missing type":        `{"specversion":"1.0","id":"i","source":"s"}`,
		"missing specversion": `{"id":"i","source":"s","type":"t"}`,
		"wrong specversion":   `{"specversion":"0.3","id":"i","source":"s","type":"t"}`,
		"invalid json":        `{`,
	} {
		if _, err := ParseStructured([]byte(body)); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestParseBinary(t *testing.T) {
	t.Parallel()

	headers := map[string]string{
		"Ce-Specversion": "1.0",
		"Ce-Id":          "evt-2",
		"Ce-Source":      "urn:example:rdw",
		"Ce-Type":        "nl.ftv.pip.entities.updated",
		"Ce-Dataversion": "v42",
		"Ce-Sequence":    "1001",
		"Content-Type":   "application/json",
	}
	body := []byte(`{"entities":[{"type":"service","id":"x"}]}`)

	e, err := Parse("application/json", headers, body)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if e.ID != "evt-2" || e.Source != "urn:example:rdw" {
		t.Errorf("unexpected context attributes: %+v", e)
	}
	if e.DataVersion() != "v42" || e.Sequence() != "1001" {
		t.Errorf("version tags not decoded: %q %q", e.DataVersion(), e.Sequence())
	}
	if string(e.Data) != string(body) {
		t.Errorf("data mismatch")
	}
	if e.DataContentType != "application/json" {
		t.Errorf("DataContentType = %q", e.DataContentType)
	}
}

func TestParseModeSelection(t *testing.T) {
	t.Parallel()

	body := []byte(`{"specversion":"1.0","id":"i","source":"s","type":"t"}`)
	e, err := Parse("application/cloudevents+json; charset=utf-8", nil, body)
	if err != nil {
		t.Fatalf("Parse structured: %v", err)
	}
	if e.ID != "i" {
		t.Errorf("structured mode not selected")
	}

	if _, err = Parse("application/json", map[string]string{}, body); err == nil {
		t.Error("binary mode without ce-* headers should fail validation")
	}
}

func TestDedup(t *testing.T) {
	t.Parallel()

	d := NewDedup(8)
	e1 := &Event{Source: "s", ID: "1"}
	e2 := &Event{Source: "s", ID: "2"}
	e3 := &Event{Source: "other", ID: "1"}

	if d.Seen(e1) {
		t.Error("first sighting must not be a duplicate")
	}
	if !d.Seen(e1) {
		t.Error("second sighting must be a duplicate")
	}
	if d.Seen(e2) || d.Seen(e3) {
		t.Error("different (source,id) tuples must not collide")
	}
}

func TestParseTraceParent(t *testing.T) {
	t.Parallel()

	trace, parent, ok := ParseTraceParent("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	if !ok || trace != "4bf92f3577b34da6a3ce929d0e0e4736" || parent != "00f067aa0ba902b7" {
		t.Errorf("unexpected result: %q %q %v", trace, parent, ok)
	}

	if _, _, ok = ParseTraceParent("garbage"); ok {
		t.Error("expected failure for malformed traceparent")
	}
}
