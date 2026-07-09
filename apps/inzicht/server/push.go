package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/cloudevents"
)

// StatisticsEventType is the CloudEvents type for pushed aggregate statistics.
const StatisticsEventType = "nl.overheid.ftv.adl.statistics"

// pushClient is overridable in tests.
var pushClient = &http.Client{Timeout: 15 * time.Second}

// StartPush launches the periodic statistics push when a positive interval and
// at least one target endpoint are configured. It returns immediately; the push
// loop runs until the context is cancelled.
func (s *Service) StartPush(ctx context.Context) {
	interval := s.cfg.Inzicht.PushInterval
	targets := s.pushTargets()
	if interval <= 0 || len(targets) == 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.pushOnce(ctx, targets)
			}
		}
	}()
}

// pushTargets returns the configured verstrekker endpoint URLs.
func (s *Service) pushTargets() []string {
	var out []string
	for _, t := range strings.Split(s.cfg.Inzicht.PushTargets, ",") {
		if t = strings.TrimSpace(t); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// pushOnce aggregates the recent window and pushes it as a CloudEvent to every
// target. Exported indirectly through StartPush; separated for testability.
func (s *Service) pushOnce(ctx context.Context, targets []string) {
	window := s.cfg.Inzicht.PushWindow
	if window <= 0 {
		window = 24 * time.Hour
	}
	to := now().UTC()
	from := to.Add(-window)

	resp, err := s.aggregate(ctx, &from, &to, "", adl.EventAccessEvaluation, s.bucket(""))
	if err != nil {
		s.logger.Warn("inzicht: statistics push aggregation failed", "error", err)
		return
	}

	event, err := s.encodeStatisticsEvent(resp)
	if err != nil {
		s.logger.Warn("inzicht: statistics push encoding failed", "error", err)
		return
	}

	for _, target := range targets {
		if err := s.postEvent(ctx, target, event); err != nil {
			s.logger.Warn("inzicht: statistics push failed", "target", target, "error", err)
			continue
		}
		s.logger.Info("inzicht: statistics pushed", "target", target, "buckets", len(resp.Statistieken))
	}
}

// encodeStatisticsEvent builds a structured-mode CloudEvents 1.0 envelope
// carrying the aggregate statistics as its data.
func (s *Service) encodeStatisticsEvent(data StatisticsResponse) ([]byte, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	envelope := map[string]any{
		"specversion":     "1.0",
		"id":              randomHex(16),
		"source":          s.cfg.Inzicht.Source,
		"type":            StatisticsEventType,
		"time":            now().UTC().Format(time.RFC3339),
		"datacontenttype": "application/json",
		"data":            json.RawMessage(payload),
	}
	return json.Marshal(envelope)
}

// postEvent delivers a structured CloudEvent to a single target endpoint.
func (s *Service) postEvent(ctx context.Context, target string, event []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(event))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", cloudevents.ContentTypeStructured)

	resp, err := pushClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return &httpStatusError{code: resp.StatusCode}
	}
	return nil
}

type httpStatusError struct{ code int }

func (e *httpStatusError) Error() string {
	return "unexpected status " + http.StatusText(e.code)
}
