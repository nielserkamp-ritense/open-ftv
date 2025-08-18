// Package pool implements a connection pool for PostgreSQL back-ends.
package pool

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pooler represents the interface for pooling PostgreSQL connections.
type Pooler interface {
	Ping(ctx context.Context) error
	Begin(ctx context.Context) (pgx.Tx, error)
	Close()
}

// NewPool instantiates a new connection pool for PostgreSQL.
func NewPool(ctx context.Context, dsn string, opts ...Option) (Pooler, error) {
	p := &Pool{maxLife: 5 * time.Minute, maxConn: 100}
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

// NewWithPooler instantiates a new connection pool for PostgreSQL using the given pooler interface.
func NewWithPooler(ctx context.Context, dsn string, pool Pooler, opts ...Option) (Pooler, error) {
	p := &Pool{maxLife: 5 * time.Minute, maxConn: 100}
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
	p.pool = pool

	return p, nil
}

// Ping tests the Postgres connection.
//
// It returns an error when the PostgreSQL backend cannot be reached.
func (p *Pool) Ping(ctx context.Context) error {
	if err := p.pool.Ping(ctx); err != nil {
		return p.pingError(err)
	}
	return nil
}

// Begin acquires a connection from the pool and begins a PostgreSQL transaction.
//
// The connection is automatically released to the pool when the transaction is committed or rolled back.
func (p *Pool) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.pool.Begin(ctx)
}

// Close cleans up and closes the PostgreSQL connection pool.
func (p *Pool) Close() {
	p.pool.Close()
}

func dsnError(err error) error {
	return fmt.Errorf("dsn parsing failed: %w", err)
}

func (p *Pool) baseError(cmd string) (string, []any) {
	return "%s:%d/%s: %s failed: ", []any{p.cfg.ConnConfig.Config.Host, p.cfg.ConnConfig.Config.Port, p.cfg.ConnConfig.Config.Database, cmd}
}

func (p *Pool) connectionError(err error) error {
	s, params := p.baseError("connection")
	return fmt.Errorf(s+" %w", append(params, err)...)
}

func (p *Pool) pingError(err error) error {
	s, params := p.baseError("ping")
	return fmt.Errorf(s+" %w", append(params, err)...)
}

// TxError can be used to generate an error related to a PostgreSQL scan failure.
func (p *Pool) TxError(err error) error {
	s, params := p.baseError("begin transaction")
	return fmt.Errorf(s+" %w", append(params, err)...)
}

// QueryError can be used to generate an error related to a PostgreSQL scan failure.
func (p *Pool) QueryError(err error, q string, params ...any) error {
	s, params2 := p.baseError("query")
	return fmt.Errorf(s+" '%s' %#v: %w", append(params2, q, params, err)...)
}

// ScanError can be used to generate an error related to a PostgreSQL scan failure.
func (p *Pool) ScanError(err error, q string, params ...any) error {
	s, params2 := p.baseError("row scan")
	return fmt.Errorf(s+" '%s' %#v: %w", append(params2, q, params, err)...)
}

type Pool struct {
	cfg     *pgxpool.Config
	pool    Pooler
	maxLife time.Duration
	maxConn int32
}
