package query

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/adl"
)

// DefaultOpenSearchSize bounds the number of hits fetched from OpenSearch in one query.
const DefaultOpenSearchSize = 5000

// OpenSearchSource reads ADL records from an OpenSearch index over the _search API.
//
// The ADL OpenSearch sink stores each record verbatim as the document (_source is the
// exact Record JSON), so reconstruction is trivially lossless: each hit's _source
// unmarshals straight back into an adl.Record.
//
// Filter pushdown: the record timestamp (a numeric field) and event_name are pushed into
// the query DSL as a range and a match filter. The remaining predicates (decision,
// afnemer, doel, trace ids) are applied client-side via Filter.Match - robust across
// dynamic field mappings, like WALSource.
type OpenSearchSource struct {
	cfg    OpenSearchConfig
	client *http.Client
}

// OpenSearchConfig configures an OpenSearchSource.
type OpenSearchConfig struct {
	// URL is the OpenSearch base URL (e.g. https://opensearch:9200).
	URL string
	// Index is the index (or alias) holding ADL records.
	Index string
	// User and Password enable HTTP basic auth (optional).
	User, Password string
	// Size bounds the number of hits fetched; defaults to DefaultOpenSearchSize.
	Size int
	// Client is the HTTP client used; defaults to a 30s-timeout client.
	Client *http.Client
}

// NewOpenSearchSource returns an OpenSearchSource for the given configuration.
func NewOpenSearchSource(cfg OpenSearchConfig) *OpenSearchSource {
	client := cfg.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	if cfg.Index == "" {
		cfg.Index = "adl"
	}
	return &OpenSearchSource{cfg: cfg, client: client}
}

// Query implements Source.
func (s *OpenSearchSource) Query(ctx context.Context, f Filter) ([]adl.Record, error) {
	size := s.cfg.Size
	if size <= 0 {
		size = DefaultOpenSearchSize
	}
	body, err := json.Marshal(openSearchQuery(f, size))
	if err != nil {
		return nil, fmt.Errorf("query: build OpenSearch query: %w", err)
	}

	endpoint := strings.TrimRight(s.cfg.URL, "/") + "/" + s.cfg.Index + "/_search"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("query: build OpenSearch request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.User != "" || s.cfg.Password != "" {
		req.SetBasicAuth(s.cfg.User, s.cfg.Password)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query: OpenSearch _search: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("query: OpenSearch _search status %d: %s", resp.StatusCode, string(raw))
	}

	var sr openSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("query: decode OpenSearch response: %w", err)
	}

	var out []adl.Record
	for _, hit := range sr.Hits.Hits {
		var rec adl.Record
		if err := json.Unmarshal(hit.Source, &rec); err != nil {
			continue
		}
		if f.Match(&rec) {
			out = append(out, rec)
		}
	}
	return out, nil
}

// openSearchQuery builds the _search body: a bool filter pushing down the timestamp
// range and event_name; the rest is filtered client-side.
func openSearchQuery(f Filter, size int) map[string]any {
	var filters []map[string]any

	if !f.From.IsZero() || !f.To.IsZero() {
		rng := map[string]any{}
		if !f.From.IsZero() {
			rng["gte"] = f.From.UnixMilli()
		}
		if !f.To.IsZero() {
			rng["lt"] = f.To.UnixMilli()
		}
		filters = append(filters, map[string]any{"range": map[string]any{"timestamp": rng}})
	}
	if f.EventName != "" {
		filters = append(filters, map[string]any{"match": map[string]any{"event_name": f.EventName}})
	}

	query := map[string]any{"match_all": map[string]any{}}
	if len(filters) > 0 {
		query = map[string]any{"bool": map[string]any{"filter": filters}}
	}
	return map[string]any{
		"size":  size,
		"sort":  []any{map[string]any{"timestamp": map[string]any{"order": "asc"}}},
		"query": query,
	}
}

type openSearchResponse struct {
	Hits struct {
		Hits []struct {
			Source json.RawMessage `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

var _ Source = (*OpenSearchSource)(nil)
