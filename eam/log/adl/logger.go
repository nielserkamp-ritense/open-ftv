package adl

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/authlog"
)

// Sink is an asynchronous destination for ADL records (OTLP, OpenSearch, ...).
type Sink interface {
	// Emit persists a single record.
	Emit(ctx context.Context, rec *Record) error
}

// Logger produces ADL records and persists them through two stages:
//
//  1. a durable, synchronous write-ahead log (JSONL append), written before the decision
//     is returned to the PEP, and
//  2. an asynchronous, idempotent flush to zero or more Sinks (OTLP and/or OpenSearch).
//
// It implements the authlog.Logger interface, so it drops in wherever the OpenSearch
// logger was used. The idempotency key is (trace_id, span_id): re-submitting the same
// record - for example when replaying the write-ahead log - does not emit a duplicate.
type Logger struct {
	cfg   Config
	info  InformationProvider
	wal   *walWriter
	sinks []Sink

	mu   sync.Mutex
	seen map[string]struct{}
}

// maxSeen bounds the in-memory idempotency set; when exceeded, the oldest knowledge is
// discarded (best-effort in-process deduplication).
const maxSeen = 1 << 16

// New instantiates an ADL Logger writing a WAL (when Config.WALPath is set) and flushing
// asynchronously to the given sinks.
func New(cfg Config, sinks ...Sink) (*Logger, error) {
	cfg.normalise()

	var wal *walWriter
	if cfg.WALPath != "" {
		var err error
		if wal, err = newWALWriter(cfg.WALPath); err != nil {
			return nil, fmt.Errorf("adl: failed to open write-ahead log %q: %w", cfg.WALPath, err)
		}
	}

	return &Logger{
		cfg:   cfg,
		info:  cfg.Information,
		wal:   wal,
		sinks: sinks,
		seen:  make(map[string]struct{}),
	}, nil
}

// Log implements the authlog.Logger interface.
//
// It builds exactly one ADL record for the evaluation (also when the PDP could not
// evaluate), writes it durably to the WAL (synchronously, before the decision goes back
// to the PEP), and then flushes it to the configured sinks. When wait is true the flush
// is synchronous; otherwise it runs in the background.
func (l *Logger) Log(ctx context.Context, wait bool, ar *authlog.AuthRecord) error {
	if ar == nil {
		return nil
	}

	return l.LogRecord(ctx, wait, l.build(ctx, ar))
}

// LogRecord durably records a pre-built ADL Record: it appends the record to the
// write-ahead log (synchronously, stage 1) and then flushes it to the configured
// sinks (stage 2, asynchronously unless wait is true). It is used for conformant
// ADL events that do not originate from a PDP evaluation - such as access to the
// Inzicht API itself - where no authlog.AuthRecord exists. The (trace_id, span_id)
// idempotency of Emit still applies.
func (l *Logger) LogRecord(ctx context.Context, wait bool, rec *Record) error {
	if rec == nil {
		return nil
	}

	// Stage 1: durable, synchronous write-ahead log.
	if l.wal != nil {
		if err := l.wal.append(rec); err != nil {
			return fmt.Errorf("adl: write-ahead log append failed: %w", err)
		}
	}

	// Stage 2: asynchronous, idempotent flush to sinks.
	if len(l.sinks) == 0 {
		return nil
	}
	if wait {
		l.Emit(ctx, rec)
		return nil
	}
	go l.Emit(context.WithoutCancel(ctx), rec)
	return nil
}

// Emit flushes a single record to all sinks. It is idempotent with respect to the
// record's (trace_id, span_id): a record that was already emitted is silently dropped,
// so redelivery (e.g. from a WAL replay) does not produce duplicates.
func (l *Logger) Emit(ctx context.Context, rec *Record) {
	if !l.markSeen(rec.IdempotencyKey()) {
		return
	}

	for _, s := range l.sinks {
		if err := s.Emit(ctx, rec); err != nil && l.cfg.Logger != nil {
			l.cfg.Logger.Error("adl: sink emit failed",
				"trace_id", rec.TraceID, "span_id", rec.SpanID, "error", err)
		}
	}
}

// Replay re-reads the write-ahead log and re-emits every record to the sinks. Thanks to
// the idempotent Emit this is safe to call at any time, e.g. on startup after a crash.
func (l *Logger) Replay(ctx context.Context) error {
	if l.cfg.WALPath == "" {
		return nil
	}

	f, err := os.Open(l.cfg.WALPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("adl: failed to open write-ahead log for replay: %w", err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		rec := &Record{}
		if err := json.Unmarshal(line, rec); err != nil {
			if l.cfg.Logger != nil {
				l.cfg.Logger.Warn("adl: skipping malformed write-ahead log line", "error", err)
			}
			continue
		}
		l.Emit(ctx, rec)
	}

	return scanner.Err()
}

// markSeen records the idempotency key and returns true if it was not seen before.
func (l *Logger) markSeen(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.seen[key]; ok {
		return false
	}
	if len(l.seen) >= maxSeen {
		l.seen = make(map[string]struct{})
	}
	l.seen[key] = struct{}{}
	return true
}

// Close releases the write-ahead log file handle.
func (l *Logger) Close() error {
	if l.wal != nil {
		return l.wal.close()
	}
	return nil
}

// compile-time assertion that Logger satisfies the authlog sink interface.
var _ authlog.Logger = (*Logger)(nil)
