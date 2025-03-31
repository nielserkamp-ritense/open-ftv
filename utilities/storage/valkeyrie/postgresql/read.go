package postgresql

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/kvtools/valkeyrie/store"
)

// Get implements the Valkeyrie Store interface.
func (db *pgDB) Get(ctx context.Context, key string, _ *store.ReadOptions) (record *store.KVPair, err error) {
	var tx pgx.Tx
	if tx, err = db.pool.Begin(ctx); err != nil {
		return nil, db.txError(err)
	}
	defer func() {
		if err == nil {
			err = tx.Commit(ctx)
		} else {
			_ = tx.Rollback(ctx)
		}
	}()

	q := strings.Replace(selectSQL, "kv", db.table, 1)

	var rows pgx.Rows
	if rows, err = tx.Query(ctx, q, key); err != nil {
		return nil, db.queryError(err, q, key)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, store.ErrKeyNotFound
	}

	if record, err = pairFromRows(rows); err != nil {
		return nil, db.scanError(err, q, key)
	}
	return
}

// List implements the Valkeyrie Store interface.
func (db *pgDB) List(ctx context.Context, directory string, _ *store.ReadOptions) (records []*store.KVPair, err error) {
	var tx pgx.Tx
	if tx, err = db.pool.Begin(ctx); err != nil {
		return nil, db.txError(err)
	}
	defer func() {
		if err == nil {
			err = tx.Commit(ctx)
		} else {
			_ = tx.Rollback(ctx)
		}
	}()

	if !strings.HasSuffix(directory, "%") {
		directory += "%"
	}

	q := strings.Replace(directorySQL, "kv", db.table, 1)

	var rows pgx.Rows
	if rows, err = tx.Query(ctx, q, directory); err != nil {
		return nil, db.queryError(err, q, directory)
	}
	defer rows.Close()

	records = make([]*store.KVPair, 0)
	for rows.Next() {
		var pair *store.KVPair
		if pair, err = pairFromRows(rows); err != nil {
			return nil, db.scanError(err, q, directory)
		}
		records = append(records, pair)
	}

	if len(records) == 0 {
		return nil, store.ErrKeyNotFound
	}
	return
}

// Exists implements the Valkeyrie Store interface.
func (db *pgDB) Exists(ctx context.Context, key string, opts *store.ReadOptions) (bool, error) {
	_, err := db.Get(ctx, key, opts)
	if err != nil {
		return false, err
	}
	return true, nil
}

func pairFromRows(rows pgx.Rows) (*store.KVPair, error) {
	out := &store.KVPair{}
	if err := rows.Scan(&out.Key, &out.LastIndex, &out.Value); err != nil {
		return nil, err
	}
	return out, nil
}

const (
	selectSQL    = `SELECT key, index, value FROM "kv" WHERE key = $1`
	directorySQL = `SELECT key, index, value FROM "kv" WHERE key LIKE $1`
)
