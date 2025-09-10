// Package postgresql implements a basic PostgreSQL database handler.
package postgresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql/pool"
)

// New instantiates a new PostgreSQL database handler with a new connection pool.
//
// The given context can be used to signal a clean shutdown of the connection pool.
func New(ctx context.Context, dsn string, maxLife time.Duration, maxConn int32) (*Postgres, error) {
	p, err := pool.NewPool(ctx, dsn, pool.WithMaxLifetime(maxLife), pool.WithMaxConnections(maxConn))
	if err != nil {
		return nil, err
	}
	return NewWithPool(ctx, dsn, maxLife, maxConn, p)
}

// NewWithPool instantiates a new PostgreSQL database handler with the given connection pool.
func NewWithPool(ctx context.Context, dsn string, maxLife time.Duration, maxConn int32, p pool.Pooler) (*Postgres, error) {
	db := &Postgres{dsn: dsn, maxLife: maxLife, maxConn: maxConn, pool: p}

	go db.background(ctx) // waits for context to be done, and closes the pool.

	return db, nil
}

// Postgres represents a PostgreSQL database connection pool.
type Postgres struct {
	dsn     string
	maxLife time.Duration
	maxConn int32
	pool    pool.Pooler
}

// ProcessRow represents the closure for handling SQL query results.
//
// The *values* parameter represents all values from a single row from the database.
type ProcessRow func(values []any) bool

// Query executes the given SQL statement with the given parameters,
// and calls the given closure for each selected row.
// The iteration process can be stopped by returning false from the closure.
func (db *Postgres) Query(ctx context.Context, q string, params []any, f ProcessRow) (err error) {
	var tx pgx.Tx
	if tx, err = db.pool.Begin(ctx); err != nil {
		return
	}

	defer func() {
		if err == nil {
			err = tx.Commit(ctx)
		} else {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	var rows pgx.Rows
	if rows, err = tx.Query(ctx, q, params...); err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var values []any
		if values, err = rows.Values(); err != nil {
			return
		}

		if !f(values) {
			break
		}
	}

	err = rows.Err()
	return
}

// Exec executes the given SQL statement with the given parameters.
func (db *Postgres) Exec(ctx context.Context, q string, params []any) (count int64, err error) {
	user := convert.AnyToString(ctx.Value("user"))
	if user == "" {
		user = "*SYSTEM*"
	}

	var tx pgx.Tx
	if tx, err = db.pool.Begin(ctx); err != nil {
		return
	}

	defer func() {
		if err == nil {
			err = tx.Commit(ctx)
		} else {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	_, err = tx.Exec(ctx, fmt.Sprintf("SELECT set_config('openftv.user', '%s', true);", user))
	if err != nil {
		return
	}

	var result pgconn.CommandTag
	if len(params) > 0 {
		result, err = tx.Exec(ctx, q, params...)
	} else {
		result, err = tx.Exec(ctx, q)
	}

	if err == nil {
		count = result.RowsAffected()
	}
	return
}

// ExecBatch executes the given set of SQL statements with the respective parameters.
func (db *Postgres) ExecBatch(ctx context.Context, statements []Statement) (count int64, err error) {
	user := convert.AnyToString(ctx.Value("user"))
	if user == "" {
		user = "*SYSTEM*"
	}

	var tx pgx.Tx
	if tx, err = db.pool.Begin(ctx); err != nil {
		return
	}

	defer func() {
		if err == nil {
			err = tx.Commit(ctx)
		} else {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	_, err = tx.Exec(ctx, fmt.Sprintf("SELECT set_config('openftv.user', '%s', true);", user))
	if err != nil {
		return
	}

	for i := range statements {
		statement := statements[i]

		var result pgconn.CommandTag
		if len(statement.Args) > 0 {
			result, err = tx.Exec(ctx, statement.SQL, statement.Args...)
		} else {
			result, err = tx.Exec(ctx, statement.SQL)
		}

		if err != nil {
			return
		}

		count += result.RowsAffected()
	}

	return
}

func (db *Postgres) background(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			db.pool.Close()
			return
		}
	}
}
