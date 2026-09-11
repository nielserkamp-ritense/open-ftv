package migrate

import (
	"context"
	"embed"
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

// PostgresEmbedded migrates a PostgreSQL database up or down by the given number of steps.
//
// If steps == 0, the function will migrate up to the highest level.
//
// The function uses the migration scripts embedded in the binary,
// and *database* as the URL for the database connection.
func PostgresEmbedded(fs embed.FS, database string, steps int, logger *slog.Logger) error {
	src := "postgresql"

	d, err := iofs.New(fs, src)
	if err != nil {
		return err
	}

	var m *migrate.Migrate
	m, err = migrate.NewWithSourceInstance(src, d, database)
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
