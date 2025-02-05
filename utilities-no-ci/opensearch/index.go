package opensearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

// Indexer represents the interface to manipulate OpenSearch indexes.
type Indexer interface {
	CreateIndex(ctx context.Context, name string, shards, replicas int) error
	DeleteIndexes(ctx context.Context, name ...string) error
}

// CreateIndex creates a new OpenSearch index.
func (l *base) CreateIndex(ctx context.Context, name string, shards, replicas int) error {
	settings := strings.NewReader(fmt.Sprintf(`{
    "settings": {
        "index": {
            "number_of_shards": %d,
            "number_of_replicas": %d
            }
        }
    }`, shards, replicas))

	req := opensearchapi.IndicesCreateRequest{
		Index: name,
		Body:  settings,
	}
	return l.checkResponse(ctx, "create index", req.Do)
}

// DeleteIndexes deletes one or more OpenSearch indexes.
func (l *base) DeleteIndexes(ctx context.Context, names ...string) error {
	req := opensearchapi.IndicesDeleteRequest{Index: names}
	return l.checkResponse(ctx, "delete index", req.Do)
}
