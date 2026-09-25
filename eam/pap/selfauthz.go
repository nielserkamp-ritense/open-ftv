package pap

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

// NewSelfAuthz builds a self-authorization PAP loaded from local (bundled) policy files: the files are the source of
// truth, so an in-memory KV store is used as its runtime index. This is not a durable persistence backend and must
// not be used for customer-managed policy data.
func NewSelfAuthz(ctx context.Context, logger *slog.Logger, language, store string, recurse bool) (*PAP, error) {
	return New(ctx, logger, WithLanguage(language), WithKeyValueDB(memory.New(), ""), WithFileStore(store, recurse))
}
