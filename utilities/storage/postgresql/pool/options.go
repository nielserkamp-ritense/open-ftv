package pool

import (
	"time"
)

// Option is the function signature for passing options when instantiating a new Postgres connection pool.
type Option func(p *Pool)

// WithMaxLifetime sets the maximum life-time of a connection.
//
// The default is 5 minutes.
func WithMaxLifetime(maxLife time.Duration) Option {
	return func(p *Pool) {
		p.maxLife = maxLife
	}
}

// WithMaxConnections sets the maximum allowed connections in the pool.
//
// The default is 100 connections.
func WithMaxConnections(m int32) Option {
	return func(p *Pool) {
		p.maxConn = m
	}
}
