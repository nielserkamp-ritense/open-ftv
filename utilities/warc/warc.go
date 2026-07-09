// Package warc implements a small, dependency-free writer and reader for the
// WARC/1.1 file format (ISO 28500).
//
// It is used by the Policy Information Point to persist an external, append-only
// log of every information exchange (outgoing pull requests and their responses,
// as well as incoming push events). Each record carries W3C-style trace and span
// identifiers in custom headers (WARC-X-ADL-Trace-Id / WARC-X-ADL-Span-Id) so an
// Authorization Decision Log can reference a stored payload as a "Logged" source
// reference, indexed on trace_id + span_id.
package warc

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// Version is the WARC format version this package writes.
const Version = "WARC/1.1"

// Record types as defined by ISO 28500.
const (
	TypeRequest  = "request"
	TypeResponse = "response"
	TypeResource = "resource"
)

// Custom (extension) header names used to link a record to an ADL trace/span.
const (
	HeaderTraceID  = "WARC-X-ADL-Trace-Id"
	HeaderSpanID   = "WARC-X-ADL-Span-Id"
	HeaderParentID = "WARC-X-ADL-Parent-Span-Id"
	HeaderVersion  = "WARC-X-ADL-Data-Version"
	HeaderSequence = "WARC-X-ADL-Sequence"
)

// Standard WARC content types for enclosed HTTP messages.
const (
	CTRequest  = "application/http;msgtype=request"
	CTResponse = "application/http;msgtype=response"
)

// Record represents a single WARC record: a set of named headers plus a content block.
type Record struct {
	Type         string            // WARC-Type (request/response/resource/...).
	RecordID     string            // WARC-Record-ID (urn:uuid:...); generated when empty.
	Date         time.Time         // WARC-Date; defaults to time.Now() when zero.
	TargetURI    string            // WARC-Target-URI (optional).
	ConcurrentTo string            // WARC-Concurrent-To (optional).
	ContentType  string            // Content-Type of the block (optional).
	Custom       map[string]string // Additional headers (e.g. WARC-X-ADL-*).
	Block        []byte            // The record payload.
}

// NewUUID returns a random RFC 4122 version-4 UUID URN (urn:uuid:...).
func NewUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4.
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10.
	return fmt.Sprintf("urn:uuid:%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// NewTraceID returns a random 16-byte trace id as 32 lowercase hex characters (W3C trace-context).
func NewTraceID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// NewSpanID returns a random 8-byte span id as 16 lowercase hex characters (W3C trace-context).
func NewSpanID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// bytesWritten serializes a record to w following ISO 28500 and returns the number of bytes written.
func (rec *Record) writeTo(w io.Writer) (int, error) {
	if rec.RecordID == "" {
		rec.RecordID = NewUUID()
	}
	if rec.Date.IsZero() {
		rec.Date = time.Now().UTC()
	}

	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "%s\r\n", Version)
	fmt.Fprintf(buf, "WARC-Type: %s\r\n", rec.Type)
	fmt.Fprintf(buf, "WARC-Record-ID: <%s>\r\n", strings.Trim(rec.RecordID, "<>"))
	fmt.Fprintf(buf, "WARC-Date: %s\r\n", rec.Date.UTC().Format(time.RFC3339Nano))

	if rec.TargetURI != "" {
		fmt.Fprintf(buf, "WARC-Target-URI: %s\r\n", rec.TargetURI)
	}
	if rec.ConcurrentTo != "" {
		fmt.Fprintf(buf, "WARC-Concurrent-To: <%s>\r\n", strings.Trim(rec.ConcurrentTo, "<>"))
	}
	if rec.ContentType != "" {
		fmt.Fprintf(buf, "Content-Type: %s\r\n", rec.ContentType)
	}

	// custom headers, written in a stable order for reproducibility.
	keys := make([]string, 0, len(rec.Custom))
	for k := range rec.Custom {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if v := rec.Custom[k]; v != "" {
			fmt.Fprintf(buf, "%s: %s\r\n", k, v)
		}
	}

	fmt.Fprintf(buf, "Content-Length: %d\r\n", len(rec.Block))
	buf.WriteString("\r\n")
	buf.Write(rec.Block)
	buf.WriteString("\r\n\r\n") // two CRLF terminate each record.

	return w.Write(buf.Bytes())
}
