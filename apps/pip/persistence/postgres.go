package persistence

import (
	"context"

	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pip/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/storage/valkeyrie/postgresql"
)

// newPG instantiates persistent storage with a Postgres backend.
func newPG(ctx context.Context, cfg *config.Config) (store.Store, error) {
	pool, err := postgresql.NewPool(
		ctx,
		cfg.PgURL,
		postgresql.WithMaxLifetime(cfg.PgMaxLife),
		postgresql.WithMaxConnections(cfg.PgMaxConn),
	)
	if err != nil {
		return nil, err
	}

	return postgresql.New(pool, cfg.PgTable)
}
