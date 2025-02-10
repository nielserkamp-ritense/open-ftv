package opensearch

import (
	"context"
	"strings"

	"github.com/defensestation/osquery"
	"github.com/goccy/go-json"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

// Searcher represents the interface to execute queries on OpenSearch.
type Searcher interface {
	SearchByQuery(ctx context.Context, index string, query *osquery.SearchRequest) (*SearchResponse, error)
	SearchBySQL(ctx context.Context, index string, query string, max int) (*SearchResponse, error)
	SearchByLucene(ctx context.Context, index string, query string, max int) (*SearchResponse, error)
}

// NewSearcher instantiates a new OpenSearch query executor.
func NewSearcher(user, pswd string, endpoints []string) (Searcher, error) {
	b, err := newBase(user, pswd, endpoints)
	if err != nil {
		return nil, err
	}
	return &search{base: *b}, nil
}

// SearchByQuery implements the Searcher interface.
//
// It can be used to execute a search with a formatted OpenSearch query.
func (s *search) SearchByQuery(ctx context.Context, index string, query *osquery.SearchRequest) (*SearchResponse, error) {
	resp, err := query.Run(
		s.client,
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(index),
	)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	out := new(SearchResponse)
	err = json.NewDecoder(resp.Body).Decode(out)
	return out, err
}

// SearchBySQL implements the Searcher interface.
//
// It can be used to execute a search with an OpenSearch SQL statement.
func (s *search) SearchBySQL(ctx context.Context, index string, query string, max int) (*SearchResponse, error) {
	req := opensearchapi.SearchRequest{Index: []string{index}, Body: strings.NewReader(query), Size: &max}

	resp, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	out := new(SearchResponse)
	err = json.NewDecoder(resp.Body).Decode(out)
	return out, err
}

// SearchByLucene implements the Searcher interface.
//
// It can be used to execute a search with a Lucene query statement.
func (s *search) SearchByLucene(ctx context.Context, index string, query string, max int) (*SearchResponse, error) {
	req := opensearchapi.SearchRequest{Index: []string{index}, Query: query, Size: &max}

	resp, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	out := new(SearchResponse)
	err = json.NewDecoder(resp.Body).Decode(out)
	return out, err
}

// SearchResponse contains the returned response for a search operation.
type SearchResponse struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []*struct {
			ID     string         `json:"_id"`
			Source map[string]any `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations map[string]any `json:"aggregations"`
}

type search struct {
	base
}
