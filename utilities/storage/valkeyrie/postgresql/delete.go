package postgresql

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/kvtools/valkeyrie/store"
)

// Delete implements the Valkeyrie Store interface.
func (db *pgDB) Delete(ctx context.Context, key string) (err error) {
	var tx pgx.Tx
	if tx, err = db.pool.Begin(ctx); err != nil {
		return db.txError(err)
	}
	defer func() {
		if err == nil {
			err = tx.Commit(ctx)
		} else {
			_ = tx.Rollback(ctx)
		}
	}()

	q := strings.Replace(deleteSQL, "kv", db.table, 1)

	if _, err = tx.Exec(ctx, q, key); err != nil {
		return db.queryError(err, q, key)
	}
	return nil
}

// AtomicDelete implements the Valkeyrie Store interface.
func (db *pgDB) AtomicDelete(ctx context.Context, key string, previous *store.KVPair) (success bool, err error) {
	var tx pgx.Tx
	if tx, err = db.pool.Begin(ctx); err != nil {
		return false, db.txError(err)
	}
	defer func() {
		if err == nil {
			err = tx.Commit(ctx)
		} else {
			_ = tx.Rollback(ctx)
		}
	}()

	q := strings.Replace(atomicDeleteSQL, "kv", db.table, 1)

	if _, err = tx.Exec(ctx, q, key, previous.LastIndex); err != nil {
		return false, db.queryError(err, q, key, previous.LastIndex)
	}
	return true, nil
}

const (
	deleteSQL       = `DELETE FROM "kv" WHERE key = $1`
	atomicDeleteSQL = `DELETE FROM "kv" WHERE key = $1 AND index = $2`
)
