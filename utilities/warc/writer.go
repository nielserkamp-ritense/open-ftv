package warc

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config holds the configuration for a rotating WARC writer.
type Config struct {
	Dir         string // directory where WARC files are stored.
	MaxBytes    int64  // rotate to a new file once the current file exceeds this size (0 = no size-based rotation).
	RotateDaily bool   // rotate to a new file when the UTC calendar day changes.
}

// Writer writes WARC records to files in a directory, rotating per day and/or per size.
//
// A Writer is safe for concurrent use.
type Writer struct {
	cfg     Config
	mutex   sync.Mutex
	file    *os.File
	name    string // basename of the current file.
	size    int64  // bytes written to the current file.
	day     string // UTC day (YYYYMMDD) of the current file.
	nowFunc func() time.Time
}

// NewWriter creates a WARC writer that stores files in cfg.Dir.
//
// The directory is created if it does not exist. Files are opened lazily on the first write.
func NewWriter(cfg Config) (*Writer, error) {
	if cfg.Dir == "" {
		return nil, fmt.Errorf("warc: no directory configured")
	}
	if err := os.MkdirAll(cfg.Dir, 0o750); err != nil {
		return nil, fmt.Errorf("warc: failed to create directory %q: %w", cfg.Dir, err)
	}
	return &Writer{cfg: cfg, nowFunc: func() time.Time { return time.Now().UTC() }}, nil
}

// Close closes the underlying file, if open.
func (w *Writer) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.file != nil {
		err := w.file.Close()
		w.file = nil
		return err
	}
	return nil
}

// Write serializes a single record and returns the basename of the file it was written to.
func (w *Writer) Write(rec *Record) (string, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if err := w.rotate(); err != nil {
		return "", err
	}

	n, err := rec.writeTo(w.file)
	w.size += int64(n)
	if err != nil {
		return w.name, err
	}
	return w.name, nil
}

// WritePair writes a request/response record pair that are mutually WARC-Concurrent-To,
// stamped with the given trace and span identifiers, and returns the file both landed in.
//
// reqBlock and respBlock are the raw HTTP request and response messages
// (see DumpRequest / DumpResponse). Either block may be nil.
func (w *Writer) WritePair(traceID, spanID, targetURI string, reqBlock, respBlock []byte, extra map[string]string) (string, error) {
	reqID := NewUUID()
	respID := NewUUID()

	custom := map[string]string{HeaderTraceID: traceID, HeaderSpanID: spanID}
	for k, v := range extra {
		custom[k] = v
	}

	req := &Record{
		Type:         TypeRequest,
		RecordID:     reqID,
		TargetURI:    targetURI,
		ConcurrentTo: respID,
		ContentType:  CTRequest,
		Custom:       cloneHeaders(custom),
		Block:        reqBlock,
	}
	resp := &Record{
		Type:         TypeResponse,
		RecordID:     respID,
		TargetURI:    targetURI,
		ConcurrentTo: reqID,
		ContentType:  CTResponse,
		Custom:       cloneHeaders(custom),
		Block:        respBlock,
	}

	name, err := w.Write(req)
	if err != nil {
		return name, err
	}
	return w.Write(resp)
}

// rotate ensures w.file points to a suitable, open file. Caller must hold the mutex.
func (w *Writer) rotate() error {
	day := w.nowFunc().Format("20060102")

	needNew := w.file == nil ||
		(w.cfg.RotateDaily && day != w.day) ||
		(w.cfg.MaxBytes > 0 && w.size >= w.cfg.MaxBytes)

	if !needNew {
		return nil
	}

	if w.file != nil {
		_ = w.file.Close()
		w.file = nil
	}

	name := fmt.Sprintf("pip-%s.warc", w.nowFunc().Format("20060102-150405.000000000"))
	f, err := os.OpenFile(filepath.Join(w.cfg.Dir, name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		return fmt.Errorf("warc: failed to open file: %w", err)
	}

	w.file = f
	w.name = name
	w.day = day
	w.size = 0
	return nil
}

func cloneHeaders(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// DumpRequest returns the raw HTTP/1.x request message (headers and, when body is true, the body).
//
// It is a thin wrapper around net/http/httputil suitable for an application/http;msgtype=request block.
func DumpRequest(req *http.Request, body bool) []byte {
	if req == nil {
		return nil
	}
	if b, err := httputil.DumpRequestOut(req, body); err == nil {
		return b
	}
	// Fall back to the server-side dump (works for inbound requests without a URL host).
	if b, err := httputil.DumpRequest(req, body); err == nil {
		return b
	}
	return nil
}

// DumpResponse returns the raw HTTP/1.x response message. It restores resp.Body so the
// caller can still read it afterwards.
func DumpResponse(resp *http.Response, body bool) []byte {
	if resp == nil {
		return nil
	}
	if b, err := httputil.DumpResponse(resp, body); err == nil {
		return b
	}
	return nil
}
