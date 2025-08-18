package postgresql

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/kvtools/valkeyrie/store"
)

// Put implements the Valkeyrie Store interface.
func (db *pgDB) Put(ctx context.Context, key string, value []byte, _ *store.WriteOptions) (err error) {
	var tx pgx.Tx
	if tx, err = db.pool.Begin(ctx); err != nil {
		return db.txError(err)
	}
	defer func() {
		if err == nil {
			err = tx.Commit(ctx)
		} else {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	q := strings.Replace(upsertSQL, "kv", db.table, -1)

	if _, err = tx.Exec(ctx, q, key, value); err != nil {
		return db.queryError(err, q, key)
	}
	return nil
}

// AtomicPut implements the Valkeyrie Store interface.
func (db *pgDB) AtomicPut(ctx context.Context, key string, value []byte, previous *store.KVPair, _ *store.WriteOptions) (success bool, next *store.KVPair, err error) {
	var tx pgx.Tx
	if tx, err = db.pool.Begin(ctx); err != nil {
		return false, nil, db.txError(err)
	}
	defer func() {
		if err == nil {
			err = tx.Commit(ctx)
		} else {
			_ = tx.Rollback(ctx)
		}
	}()

	var q string
	var params []any

	if previous == nil {
		q = strings.Replace(insertSQL, "kv", db.table, -1)
		params = []any{key, value}
		next = &store.KVPair{Key: key, Value: value, LastIndex: 1}
	} else {
		q = strings.Replace(updateSQL, "kv", db.table, -1)
		params = []any{value, key, previous.LastIndex}
		next = &store.KVPair{Key: key, Value: value, LastIndex: previous.LastIndex + 1}
	}

	if _, err = tx.Exec(ctx, q, params...); err != nil {
		return false, nil, db.queryError(err, q, key)
	}
	return true, next, nil
}

// upsert statement; new index == 1, otherwise += 1.
const upsertSQL = `
INSERT INTO "kv" (key, index, value) VALUES ($1, 1, $2)
 ON CONFLICT (key) DO UPDATE SET index = kv.index + 1, value = EXCLUDED.value`

// insert statement; initial index == 1.
const insertSQL = `
INSERT INTO "kv" (key, index, value) VALUES ($1, 1, $2)`

// update statement; index += 1.
const updateSQL = `
UPDATE "kv" SET index = index + 1, value = $1 WHERE key = $2 AND index = $3`
