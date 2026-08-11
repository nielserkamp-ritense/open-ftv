package decisions

import (
	"errors"
	"log/slog"
	"path/filepath"
	"strings"

	migrate2 "github.com/golang-migrate/migrate/v4"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions/migrations"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/migrate"
)

// Migrate brings the Authorization Decision Log schema in the database at *dsn* up to date.
//
// The schema belongs to this component, not to a single app: every app configured with a
// PostgreSQL ADL migrates it at startup, because no app is present in every deployment.
// Simultaneous migration is safe; golang-migrate takes an advisory lock.
//
// The function is a no-op without a migration source, or when neither auto nor a number of
// steps is requested. If auto is set, steps is ignored and the database is migrated up to
// the latest level; a negative number of steps migrates down.
func Migrate(source, dsn string, steps int, auto bool, logger *slog.Logger) (err error) {
	if source == "" || (steps == 0 && !auto) {
		return nil
	}

	if auto {
		steps = 0
	}

	switch {
	case strings.EqualFold(source, "*embed*"):
		err = migrate.PostgresEmbedded(migrations.PgDecisionLog, dsn, steps, logger)

	default:
		if !strings.HasPrefix(source, "file://") {
			source, _ = filepath.Abs(source)
			source = "file://" + source
		}

		err = migrate.Postgres(source, dsn, steps, logger)
	}

	if err != nil && !errors.Is(err, migrate2.ErrNoChange) {
		logger.Error("decision log database migration failed", "auto", auto, "steps", steps, "err", err)
		return err
	}

	if err == nil {
		logger.Info("decision log database migration completed successfully")
	} else {
		logger.Info("decision log database migration completed; no changes")
	}

	return nil
}
