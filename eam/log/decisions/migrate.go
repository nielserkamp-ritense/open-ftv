package decisions

import (
	"errors"
	"log/slog"

	migrate2 "github.com/golang-migrate/migrate/v4"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions/migrations"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/migrate"
)

// Migrate brings the Authorization Decision Log schema in the database at *dsn* up to the latest
// level, using the migration scripts embedded in the binary.
//
// The schema belongs to this component, not to a single app: every app that uses the ADL migrates
// it at startup, because no app is present in every deployment. Simultaneous migration is safe;
// golang-migrate takes an advisory lock.
func Migrate(dsn string, logger *slog.Logger) error {
	err := migrate.PostgresEmbedded(migrations.PgDecisionLog, dsn, 0, logger)
	if err != nil && !errors.Is(err, migrate2.ErrNoChange) {
		logger.Error("decision log database migration failed", "err", err)
		return err
	}

	if err == nil {
		logger.Info("decision log database migration completed successfully")
	} else {
		logger.Info("decision log database migration completed; no changes")
	}

	return nil
}
