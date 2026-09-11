package pap

import (
	"errors"

	migrate2 "github.com/golang-migrate/migrate/v4"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/migrations"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/migrate"
)

// Migrate brings the database configured via WithMigration up to the latest level, using the
// migration scripts embedded in the binary. It is a no-op when no database is configured.
func (p *PAP) Migrate() error {
	if p.migrateDB == "" {
		return nil
	}

	err := migrate.PostgresEmbedded(migrations.PostgreSQL, p.migrateDB, 0, p.logger)
	if err != nil && !errors.Is(err, migrate2.ErrNoChange) {
		p.logger.Error("pap database migration failed", "err", err)
		return err
	}

	if err == nil {
		p.logger.Info("pap database migration completed successfully")
	} else {
		p.logger.Info("pap database migration completed; no changes")
	}

	return nil
}
