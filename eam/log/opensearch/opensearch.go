package opensearch

import (
	"context"
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities-no-ci/opensearch"
)

// NewOpenSearch instantiates a new OpenSearch sink for the authorisation log.
func NewOpenSearch(index, user, pswd string, endpoints ...string) (Logger, error) {
	sink, err := newLogger(user, pswd, endpoints)
	if err != nil {
		return nil, fmt.Errorf("failed to create new logger: %v", err)
	}
	return &os{index: index, sink: sink}, nil
}

// Log implements the Logger interface.
func (l *os) Log(ctx context.Context, wait bool, record *AuthRecord) error {
	return l.sink.Log(ctx, wait, opensearch.LogRecord{Index: l.index, Data: record})
}

type os struct {
	index string
	sink  opensearch.Logger
}

var newLogger = opensearch.NewLogger // override in unit tests.
