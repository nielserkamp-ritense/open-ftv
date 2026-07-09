package adl

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// walWriter is a durable, append-only JSONL write-ahead log. Each record is written as a
// single JSON line and fsync'd before append returns, so the record is on stable storage
// before the decision is handed back to the PEP.
type walWriter struct {
	mu sync.Mutex
	f  *os.File
}

func newWALWriter(path string) (*walWriter, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return nil, err
		}
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		return nil, err
	}
	return &walWriter{f: f}, nil
}

// append serialises the record as one JSON line and flushes it to stable storage.
func (w *walWriter) append(rec *Record) error {
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, err := w.f.Write(data); err != nil {
		return err
	}
	return w.f.Sync()
}

func (w *walWriter) close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.f.Close()
}
