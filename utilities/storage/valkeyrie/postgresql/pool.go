package postgresql

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pooler represents the interface for pooling Postgres connections.
type Pooler interface {
	Ping(ctx context.Context) error
	Begin(ctx context.Context) (pgx.Tx, error)
	Close()
}

// NewPool instantiates a new connection pool for Postgres.
func NewPool(ctx context.Context, dsn string, opts ...Option) (Pooler, error) {
	p := &pool{maxLife: 5 * time.Minute, maxConn: 100}
	for i := range opts {
		opts[i](p)
	}

	var err error
	p.cfg, err = pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, dsnError(err)
	}

	p.cfg.MaxConnLifetime = p.maxLife
	p.cfg.MaxConns = p.maxConn

	if p.pool, err = pgxpool.NewWithConfig(ctx, p.cfg); err != nil {
		return nil, p.connectionError(err)
	}
	return p, nil
}

// Ping tests the Postgres connection.
//
// It returns an error when the Postgres backend cannot be reached.
func (p *pool) Ping(ctx context.Context) error {
	if err := p.pool.Ping(ctx); err != nil {
		return p.pingError(err)
	}
	return nil
}

// Begin acquires a connection from the pool and begins a Postgres transaction.
//
// The connection is automatically released to the pool when the transaction is committed or rolled back.
func (p *pool) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.pool.Begin(ctx)
}

// Close cleans up and closes the Postgres connection pool.
func (p *pool) Close() {
	p.pool.Close()
}

type pool struct {
	cfg     *pgxpool.Config
	pool    Pooler
	maxLife time.Duration
	maxConn int32
}
