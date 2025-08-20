package pip

import (
	"context"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

// NewPostgresDB instantiates a new PostgreSQL database connection handler.
//
// The given context is used to signal a clean shutdown of the connection pool.
func NewPostgresDB(ctx context.Context, dsn string, maxLife time.Duration, maxConn int32) (*PostgresDB, error) {
	p, err := postgresql.New(ctx, dsn, maxLife, maxConn)
	if err != nil {
		return nil, err
	}
	return &PostgresDB{p: p, now: time.Now}, nil
}

// NewPostgresWithPool instantiates a new PostgreSQL database connection handler using the given connection pool.
func NewPostgresWithPool(pool *postgresql.Postgres) *PostgresDB {
	return &PostgresDB{p: pool, now: time.Now}
}

// PostgresDB wraps a PostgreSQL connection pool with attribute, entity and relation management functions.
type PostgresDB struct {
	p   *postgresql.Postgres
	now func() time.Time // for time-sensitive unit-tests.
}

func timeToLastIndex(t time.Time) uint64 {
	if t.IsZero() {
		return 0
	}
	return uint64(t.UnixNano())
}

func timeFromLastIndex(lastIndex uint64) time.Time {
	if lastIndex == 0 {
		return time.Time{}
	}
	i := int64(lastIndex)
	j := int64(time.Second)
	return time.Unix(i/j, i%j).UTC()
}
