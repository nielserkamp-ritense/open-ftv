package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// Record is the minimal ADL record shape the replayer reads: the body (raw
// request/response/configuration payloads) and the attributes (source references,
// including adl.core.policies).
type Record struct {
	TraceID    string         `json:"trace_id"`
	SpanID     string         `json:"span_id"`
	EventName  string         `json:"event_name"`
	Status     string         `json:"status"`
	Attributes map[string]any `json:"attributes"`
	Body       map[string]any `json:"body"`
}

// ReadRecord reads a single ADL record from a file. The file may be a JSON document or
// a JSONL write-ahead log; the first non-empty line is used.
func ReadRecord(path string) (*Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		rec := &Record{}
		if err := json.Unmarshal(line, rec); err != nil {
			return nil, fmt.Errorf("parse record: %w", err)
		}
		return rec, nil
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("no record found in %s", path)
}
