package opensearch

import (
	"bytes"
	"context"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

// LogRecord represents a basic log record.
type LogRecord struct {
	Index string `json:"index"`
	ID    string `json:"id,omitempty"`
	Data  any    `json:"data"`
}

// Logger represents the interface to log records in OpenSearch.
type Logger interface {
	Indexer
	Log(ctx context.Context, rec LogRecord) error
	LogBulk(ctx context.Context, records ...LogRecord) error
}

// NewLogger instantiates a new OpenSearch logger.
func NewLogger(user, pswd string, endpoints []string) (Logger, error) {
	b, err := newBase(user, pswd, endpoints)
	if err != nil {
		return nil, err
	}
	return &logger{base: *b}, nil
}

// Log writes the given record to the given OpenSearch index.
//
// If the ID field in the record is empty, a new uuid will be generated for it.
func (l *logger) Log(ctx context.Context, rec LogRecord) error {
	if rec.ID == "" {
		rec.ID = newID()
	}

	b, err := json.Marshal(rec.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req := opensearchapi.IndexRequest{Index: rec.Index, DocumentID: rec.ID, Body: bytes.NewReader(b)}
	return l.checkResponse(ctx, "write log", req.Do)
}

// LogBulk writes the given records to the given OpenSearch index, using a bulk operation.
//
// If the ID field in a record is empty, a new uuid will be generated for it.
func (l *logger) LogBulk(ctx context.Context, records ...LogRecord) error {
	var buf bytes.Buffer

	for i := range records {
		rec := records[i]
		if rec.ID == "" {
			rec.ID = newID()
		}

		b, err := json.Marshal(createLog{Create: createLogData{Index: rec.Index, ID: rec.ID}})
		if err != nil {
			return fmt.Errorf("failed to marshal action: %w", err)
		}

		buf.Write(b)
		buf.WriteByte('\n')

		b, err = json.Marshal(rec)
		if err != nil {
			return fmt.Errorf("failed to marshal action: %w", err)
		}

		buf.Write(b)
		buf.WriteByte('\n')
	}

	req := opensearchapi.BulkRequest{Body: bytes.NewReader(buf.Bytes())}
	return l.checkResponse(ctx, "bulk-write log", req.Do)
}

type logger struct {
	base
}

type createLog struct {
	Create createLogData `json:"create"`
}

type createLogData struct {
	Index string `json:"_index"`
	ID    string `json:"_id"`
}
