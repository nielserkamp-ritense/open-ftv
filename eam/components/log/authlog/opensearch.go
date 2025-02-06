package authlog

import (
	"context"
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities-no-ci/opensearch"
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
func (l *os) Log(ctx context.Context, record *AuthRecord) error {
	return l.sink.Log(ctx, opensearch.LogRecord{Index: l.index, Data: record})
}

type os struct {
	index string
	sink  opensearch.Logger
}

var newLogger = opensearch.NewLogger // override in unit tests.
