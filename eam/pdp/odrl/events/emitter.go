// Package events implements the near-real-time distribution side of the ODRL
// im-/export: a models.EventSink that, on every local policy change in the
// PAP, renders the affected policy as ODRL-AP-NL and
//
//   - pushes it as a CloudEvents 1.0 event (structured mode, at-least-once
//     with retry/backoff) to a configured list of subscriber endpoints, and
//   - optionally writes it to an export directory (file-based distribution).
package events

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/export"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/cloudevents"
	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
)

// CloudEvents event types emitted on policy changes.
const (
	EventTypeUpdated = "nl.overheid.ftv.policy.updated"
	EventTypeRemoved = "nl.overheid.ftv.policy.removed"
)

// Emitter pushes policy changes as CloudEvents and/or exports them as files.
// It implements models.EventSink and is registered on the PAP.
type Emitter struct {
	ctx         context.Context
	logger      *slog.Logger
	pap         pap.PAP
	exporter    *export.Exporter
	subscribers []string
	exportDir   string
	source      string
	client      *http.Client
	retries     int
	backoff     time.Duration
	wg          sync.WaitGroup
}

// Option customizes an Emitter.
type Option func(*Emitter)

// WithClient overrides the HTTP client (useful for tests).
func WithClient(c *http.Client) Option { return func(e *Emitter) { e.client = c } }

// WithRetry configures at-least-once delivery: attempts (>=1) and the initial
// backoff, which doubles after every failed attempt.
func WithRetry(attempts int, backoff time.Duration) Option {
	return func(e *Emitter) {
		if attempts > 0 {
			e.retries = attempts
		}
		if backoff > 0 {
			e.backoff = backoff
		}
	}
}

// New creates an Emitter.
//
//   - subscribers: CloudEvents endpoint URLs to push to (may be empty).
//   - exportDir: directory for file export (empty disables file export).
//   - source: value for the CloudEvents "source" attribute.
func New(ctx context.Context, logger *slog.Logger, p pap.PAP, exp *export.Exporter,
	subscribers []string, exportDir, source string, options ...Option) *Emitter {
	if ctx == nil {
		ctx = context.Background()
	}
	if logger == nil {
		logger = slog.Default()
	}
	if source == "" {
		source = "urn:ftv:pap"
	}

	e := &Emitter{
		ctx:         ctx,
		logger:      logger,
		pap:         p,
		exporter:    exp,
		subscribers: subscribers,
		exportDir:   exportDir,
		source:      source,
		client:      &http.Client{Timeout: 10 * time.Second},
		retries:     4,
		backoff:     250 * time.Millisecond,
	}
	for i := range options {
		options[i](e)
	}
	return e
}

// Handle implements the models.EventSink interface. It never blocks the PAP:
// rendering and delivery happen on a separate goroutine.
func (e *Emitter) Handle(t models.EventType, key string) {
	switch t {
	case models.PolicyAdded, models.PolicyReplaced, models.PolicyRemoved:
	default:
		return // not a policy event.
	}

	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		e.process(t, key)
	}()
}

// Wait blocks until all in-flight deliveries are done (used by tests and
// orderly shutdown).
func (e *Emitter) Wait() { e.wg.Wait() }

func (e *Emitter) process(t models.EventType, key string) {
	eventType := EventTypeUpdated
	var payload []byte

	if t == models.PolicyRemoved {
		eventType = EventTypeRemoved
	} else {
		language, id := splitKey(key)
		doc, err := e.exporter.ExportPolicy(language, id)
		if err != nil {
			e.logger.Warn("odrl events: cannot export policy", "key", key, "err", err)
		} else {
			var buf bytes.Buffer
			if err = doc.Serialize(&buf, mime.MimeTypeTurtle); err != nil {
				e.logger.Warn("odrl events: cannot serialize policy", "key", key, "err", err)
			} else {
				payload = buf.Bytes()
			}
		}
	}

	e.exportFile(t, key, payload)

	if len(e.subscribers) == 0 {
		return
	}

	event, err := e.buildEvent(eventType, key, payload)
	if err != nil {
		e.logger.Error("odrl events: cannot build event", "key", key, "err", err)
		return
	}

	for _, sub := range e.subscribers {
		e.deliver(sub, event)
	}
}

// buildEvent renders a CloudEvents 1.0 structured-mode envelope. The ODRL
// document travels as data_base64 (Turtle is not JSON); the sha256 of the
// payload is exposed as the "dataversion" extension.
func (e *Emitter) buildEvent(eventType, key string, payload []byte) ([]byte, error) {
	event := map[string]any{
		"specversion": "1.0",
		"id":          uuid.NewString(),
		"source":      e.source,
		"type":        eventType,
		"subject":     key,
		"time":        time.Now().UTC().Format(time.RFC3339),
	}

	if payload != nil {
		sum := sha256.Sum256(payload)
		event["datacontenttype"] = mime.MimeTypeTurtle
		event["data_base64"] = base64.StdEncoding.EncodeToString(payload)
		event[cloudevents.ExtDataVersion] = hex.EncodeToString(sum[:])
	}

	return json.Marshal(event)
}

// deliver pushes one event to one subscriber, retrying with exponential
// backoff (at-least-once semantics).
func (e *Emitter) deliver(url string, event []byte) {
	backoff := e.backoff

	for attempt := 1; attempt <= e.retries; attempt++ {
		if err := e.post(url, event); err == nil {
			return
		} else if attempt == e.retries {
			e.logger.Error("odrl events: delivery failed, giving up", "subscriber", url, "attempts", attempt, "err", err)
			return
		} else {
			e.logger.Warn("odrl events: delivery failed, will retry", "subscriber", url, "attempt", attempt, "err", err)
		}

		select {
		case <-e.ctx.Done():
			return
		case <-time.After(backoff):
			backoff *= 2
		}
	}
}

func (e *Emitter) post(url string, event []byte) error {
	req, err := http.NewRequestWithContext(e.ctx, http.MethodPost, url, bytes.NewReader(event))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", cloudevents.ContentTypeStructured)

	resp, err2 := e.client.Do(req)
	if err2 != nil {
		return err2
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

// exportFile maintains the file-based export: one Turtle file per policy key.
func (e *Emitter) exportFile(t models.EventType, key string, payload []byte) {
	if e.exportDir == "" {
		return
	}

	path := filepath.Join(e.exportDir, filepath.FromSlash(sanitizeKey(key))+".ttl")

	if t == models.PolicyRemoved {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			e.logger.Warn("odrl events: cannot remove export file", "path", path, "err", err)
		}
		return
	}

	if payload == nil {
		return
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		e.logger.Warn("odrl events: cannot create export dir", "path", path, "err", err)
		return
	}
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		e.logger.Warn("odrl events: cannot write export file", "path", path, "err", err)
	}
}

// splitKey splits a PAP policy key into language (the first segment) and id
// (the rest, which may itself contain separators).
func splitKey(key string) (string, string) {
	if i := strings.IndexByte(key, '/'); i > 0 {
		return key[:i], key[i+1:]
	}
	return "", key
}

// sanitizeKey keeps policy keys inside the export directory.
func sanitizeKey(key string) string {
	parts := strings.Split(key, "/")
	out := parts[:0]
	for _, p := range parts {
		if p == "" || p == "." || p == ".." {
			continue
		}
		out = append(out, p)
	}
	return strings.Join(out, "/")
}
