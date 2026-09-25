package pip

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

// NewSelfAuthz builds a self-authorization PIP loaded from local attribute/entity files.
// The files are the source of truth. An in-memory KV store is used as its runtime index.
func NewSelfAuthz(ctx context.Context, logger *slog.Logger, store string, recurse bool, pullConfigs string) (*PIP, error) {
	opts := []Option{
		WithKeyValueDB(memory.New(), ""),
		WithFileStore(store, recurse),
	}

	if pullConfigs != "" {
		opts = append(opts, WithPullConfigs(pullConfigs))
	}

	return New(ctx, logger, opts...)
}
