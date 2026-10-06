package migrate

import (
	"context"
	"embed"
	"errors"
	"log/slog"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// PostgresEmbedded migrates a PostgreSQL database up or down by the given number of steps.
//
// If steps == 0, the function will migrate up to the highest level.
//
// The function uses the migration scripts embedded in the binary,
// and *database* as the URL for the database connection.
func PostgresEmbedded(fs embed.FS, database string, steps int, logger *slog.Logger) error {
	u, err := url.Parse(database)
	if err != nil {
		return err
	}

	// golang-migrate registers its pgx driver under its own scheme.
	u.Scheme = "pgx5"

	src := "postgresql"

	d, err := iofs.New(fs, src)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance(src, d, u.String())
	if err != nil {
		return err
	}

	return doPostgres(m, steps, logger)
}

func doPostgres(m *migrate.Migrate, steps int, logger *slog.Logger) (err error) {
	defer func() {
		e1, e2 := m.Close()
		err = errors.Join(err, e1, e2)
	}()

	m.Log = &wrapper{logger: logger, verbose: logger.Enabled(context.Background(), slog.LevelDebug)}

	switch {
	case steps == 0:
		return m.Up()
	default:
		return m.Steps(steps)
	}
}
