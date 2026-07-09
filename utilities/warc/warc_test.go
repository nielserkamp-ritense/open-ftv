package warc

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRecordRoundTrip(t *testing.T) {
	t.Parallel()

	trace := NewTraceID()
	span := NewSpanID()

	orig := &Record{
		Type:         TypeRequest,
		Date:         time.Date(2026, 7, 8, 10, 30, 0, 0, time.UTC),
		TargetURI:    "https://source.example/v1/persons/123",
		ConcurrentTo: NewUUID(),
		ContentType:  CTRequest,
		Custom: map[string]string{
			HeaderTraceID:  trace,
			HeaderSpanID:   span,
			HeaderVersion:  "1.4.0",
			HeaderSequence: "42",
		},
		Block: []byte("GET /v1/persons/123 HTTP/1.1\r\nHost: source.example\r\n\r\n"),
	}

	buf := &bytes.Buffer{}
	if _, err := orig.writeTo(buf); err != nil {
		t.Fatalf("writeTo: %v", err)
	}

	recs, err := Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}

	got := recs[0]
	if got.Type != orig.Type {
		t.Errorf("Type = %q, want %q", got.Type, orig.Type)
	}
	if got.RecordID != strings.Trim(orig.RecordID, "<>") {
		t.Errorf("RecordID = %q, want %q", got.RecordID, orig.RecordID)
	}
	if !got.Date.Equal(orig.Date) {
		t.Errorf("Date = %v, want %v", got.Date, orig.Date)
	}
	if got.TargetURI != orig.TargetURI {
		t.Errorf("TargetURI = %q, want %q", got.TargetURI, orig.TargetURI)
	}
	if got.ConcurrentTo != orig.ConcurrentTo {
		t.Errorf("ConcurrentTo = %q, want %q", got.ConcurrentTo, orig.ConcurrentTo)
	}
	if got.ContentType != orig.ContentType {
		t.Errorf("ContentType = %q, want %q", got.ContentType, orig.ContentType)
	}
	if !bytes.Equal(got.Block, orig.Block) {
		t.Errorf("Block = %q, want %q", got.Block, orig.Block)
	}
	if got.Custom[HeaderTraceID] != trace {
		t.Errorf("trace header = %q, want %q", got.Custom[HeaderTraceID], trace)
	}
	if got.Custom[HeaderSpanID] != span {
		t.Errorf("span header = %q, want %q", got.Custom[HeaderSpanID], span)
	}
	if got.Custom[HeaderVersion] != "1.4.0" {
		t.Errorf("version header = %q, want %q", got.Custom[HeaderVersion], "1.4.0")
	}
}

func TestWritePairAndRead(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w, err := NewWriter(Config{Dir: dir})
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	defer w.Close()

	trace, span := NewTraceID(), NewSpanID()
	reqBlock := []byte("GET /a HTTP/1.1\r\nHost: x\r\n\r\n")
	respBlock := []byte("HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok")

	name, err := w.WritePair(trace, span, "https://x/a", reqBlock, respBlock, map[string]string{HeaderVersion: "9"})
	if err != nil {
		t.Fatalf("WritePair: %v", err)
	}
	if name == "" {
		t.Fatal("expected a file name")
	}

	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	recs, err := Read(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("expected 2 records, got %d", len(recs))
	}
	if recs[0].Type != TypeRequest || recs[1].Type != TypeResponse {
		t.Fatalf("unexpected record types: %s, %s", recs[0].Type, recs[1].Type)
	}
	// Concurrent-To must cross-reference the pair.
	if recs[0].ConcurrentTo != recs[1].RecordID || recs[1].ConcurrentTo != recs[0].RecordID {
		t.Errorf("Concurrent-To cross reference broken")
	}
	if recs[0].Custom[HeaderTraceID] != trace || recs[1].Custom[HeaderSpanID] != span {
		t.Errorf("trace/span headers not propagated to both records")
	}
	if recs[0].Custom[HeaderVersion] != "9" {
		t.Errorf("extra header not propagated: %q", recs[0].Custom[HeaderVersion])
	}
}

func TestRotation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w, err := NewWriter(Config{Dir: dir, MaxBytes: 10})
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	defer w.Close()

	names := map[string]bool{}
	for i := 0; i < 3; i++ {
		n, err := w.Write(&Record{Type: TypeResource, Block: []byte("0123456789ABCDEF")})
		if err != nil {
			t.Fatalf("Write: %v", err)
		}
		names[n] = true
		time.Sleep(2 * time.Millisecond) // ensure a distinct timestamp-based filename.
	}
	if len(names) < 2 {
		t.Errorf("expected size-based rotation to produce multiple files, got %d", len(names))
	}
}

func TestDumpRoundTrip(t *testing.T) {
	t.Parallel()

	req, _ := http.NewRequest(http.MethodPost, "https://src.example/v1/events", strings.NewReader(`{"a":1}`))
	req.Header.Set("Content-Type", "application/json")

	block := DumpRequest(req, true)
	if !bytes.Contains(block, []byte("POST /v1/events HTTP/1.1")) {
		t.Errorf("dump missing request line: %q", block)
	}
	if !bytes.Contains(block, []byte(`{"a":1}`)) {
		t.Errorf("dump missing body: %q", block)
	}
}
