package query

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
)

// DefaultLokiLookback is the time window queried when a Filter carries no From bound.
// Loki's query_range API requires an explicit start; we default to 90 days so that
// unbounded queries still return recent history without scanning the whole store.
const DefaultLokiLookback = 90 * 24 * time.Hour

// DefaultLokiLimit bounds the number of entries fetched from Loki in one query.
const DefaultLokiLimit = 5000

// LokiSource reads ADL records from a Grafana Loki backend over the HTTP
// query_range API. It is a drop-in replacement for WALSource: the collector's Loki
// exporter stores each ADL record as one log line whose text is the record JSON
// (identical to the write-ahead-log JSONL form), so a line decodes straight back into
// an adl.Record - a lossless round-trip.
//
// Filter pushdown: the bounded-cardinality labels (event_name, decision) and the time
// window are pushed into the LogQL stream selector and query_range start/end, so Loki
// only returns the relevant streams. The remaining predicates (afnemer, doel, trace ids)
// are applied client-side via Filter.Match, exactly like WALSource.
type LokiSource struct {
	cfg    LokiConfig
	client *http.Client
}

// LokiConfig configures a LokiSource.
type LokiConfig struct {
	// URL is the Loki base URL (e.g. http://loki:3100). The query_range path is appended.
	URL string
	// OrgID sets the X-Scope-OrgID header for multi-tenant Loki (optional).
	OrgID string
	// User and Password enable HTTP basic auth (optional).
	User, Password string
	// Limit bounds the number of entries fetched; defaults to DefaultLokiLimit.
	Limit int
	// Lookback is the window queried when Filter.From is zero; defaults to DefaultLokiLookback.
	Lookback time.Duration
	// Client is the HTTP client used; defaults to a 30s-timeout client.
	Client *http.Client
}

// NewLokiSource returns a LokiSource for the given configuration.
func NewLokiSource(cfg LokiConfig) *LokiSource {
	client := cfg.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &LokiSource{cfg: cfg, client: client}
}

// Query implements Source.
func (s *LokiSource) Query(ctx context.Context, f Filter) ([]adl.Record, error) {
	start, end := s.window(f)

	endpoint := strings.TrimRight(s.cfg.URL, "/") + "/loki/api/v1/query_range"
	q := url.Values{}
	q.Set("query", lokiSelector(f))
	q.Set("start", strconv.FormatInt(start.UnixNano(), 10))
	q.Set("end", strconv.FormatInt(end.UnixNano(), 10))
	q.Set("direction", "forward")
	limit := s.cfg.Limit
	if limit <= 0 {
		limit = DefaultLokiLimit
	}
	q.Set("limit", strconv.Itoa(limit))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("query: build Loki request: %w", err)
	}
	if s.cfg.OrgID != "" {
		req.Header.Set("X-Scope-OrgID", s.cfg.OrgID)
	}
	if s.cfg.User != "" || s.cfg.Password != "" {
		req.SetBasicAuth(s.cfg.User, s.cfg.Password)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query: Loki query_range: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("query: Loki query_range status %d: %s", resp.StatusCode, string(body))
	}

	var lr lokiResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, fmt.Errorf("query: decode Loki response: %w", err)
	}

	var out []adl.Record
	for _, stream := range lr.Data.Result {
		for _, entry := range stream.Values {
			if len(entry) < 2 {
				continue
			}
			var rec adl.Record
			if err := json.Unmarshal([]byte(entry[1]), &rec); err != nil {
				continue // skip malformed lines rather than abort, like WALSource.
			}
			if f.Match(&rec) {
				out = append(out, rec)
			}
		}
	}
	return out, nil
}

// window resolves the query_range [start, end) from the filter, applying the lookback
// default when From is zero and now() when To is zero.
func (s *LokiSource) window(f Filter) (time.Time, time.Time) {
	end := f.To
	if end.IsZero() {
		end = time.Now()
	}
	start := f.From
	if start.IsZero() {
		lookback := s.cfg.Lookback
		if lookback <= 0 {
			lookback = DefaultLokiLookback
		}
		start = end.Add(-lookback)
	}
	return start, end
}

// lokiSelector builds the LogQL stream selector from the pushed-down labels. The
// constant job="adl" matcher is always present so the selector is never empty.
func lokiSelector(f Filter) string {
	matchers := []string{fmt.Sprintf("%s=%q", adl.LabelJob, adl.JobValue)}
	if f.EventName != "" {
		matchers = append(matchers, fmt.Sprintf("%s=%q", adl.LabelEventName, f.EventName))
	}
	if f.Decision != nil {
		val := adl.DecisionDeny
		if *f.Decision {
			val = adl.DecisionPermit
		}
		matchers = append(matchers, fmt.Sprintf("%s=%q", adl.LabelDecision, val))
	}
	return "{" + strings.Join(matchers, ", ") + "}"
}

// lokiResponse models the subset of the Loki query_range JSON response we use.
type lokiResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

var _ Source = (*LokiSource)(nil)
