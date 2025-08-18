package migrate

import (
	"context"
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/migrations"
)

// Postgres migrates a PostgreSQL database up or down by the given number of steps.
//
// If steps == 0, the function will migrate up to the highest level.
//
// The function uses the migration scripts located by *source* (file system)
// and *database* as the URL for the database connection.
func Postgres(src, database string, steps int, logger *slog.Logger) error {
	m, err := migrate.New(src, database)
	if err != nil {
		return err
	}
	return doPostgres(m, steps, logger)
}

// PostgresEmbedded migrates a PostgreSQL database up or down by the given number of steps.
//
// If steps == 0, the function will migrate up to the highest level.
//
// The function uses the migration scripts embedded in the binary,
// and *database* as the URL for the database connection.
func PostgresEmbedded(database string, steps int, logger *slog.Logger) error {
	src := "postgresql"

	d, err := iofs.New(migrations.PostgreSQL, src)
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
