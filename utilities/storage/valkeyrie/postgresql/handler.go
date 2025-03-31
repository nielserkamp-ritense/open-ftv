package postgresql

import (
	"context"
	"fmt"

	"github.com/kvtools/valkeyrie/store"
	_ "github.com/lib/pq"
)

// New instantiates a Valkeyrie Store using the pgx library to connect to a Postgres database as the storage engine.
//
// See https://github.com/kvtools/valkeyrie.
//
// The dsn parameter should represent a valid Postgres instance with appropriate credentials.
// The table parameter should point to a table within the Postgres schema as defined by ./sql/create.sql.
func New(pool Pooler, table string) (store.Store, error) {
	return &pgDB{pool: pool, table: table}, nil
}

// Close implements the Valkeyrie Store interface.
func (db *pgDB) Close() error {
	db.pool.Close()
	return nil
}

// Watch implements the Valkeyrie Store interface.
func (db *pgDB) Watch(_ context.Context, _ string, _ *store.ReadOptions) (<-chan *store.KVPair, error) {
	return nil, ni
}

// WatchTree implements the Valkeyrie Store interface.
func (db *pgDB) WatchTree(_ context.Context, _ string, _ *store.ReadOptions) (<-chan []*store.KVPair, error) {
	return nil, ni
}

// DeleteTree implements the Valkeyrie Store interface.
func (db *pgDB) DeleteTree(_ context.Context, _ string) error {
	return ni
}

// NewLock implements the Valkeyrie Store interface.
func (db *pgDB) NewLock(_ context.Context, _ string, _ *store.LockOptions) (store.Locker, error) {
	return nil, ni
}

type pgDB struct {
	pool  Pooler
	table string
}

var ni = fmt.Errorf("not implemented")
