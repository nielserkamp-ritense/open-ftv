package pap

import (
	"errors"
	"path/filepath"
	"strings"

	migrate2 "github.com/golang-migrate/migrate/v4"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/migrations"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/migrate"
)

// migrationConfigured reports whether Source, DB, and either Auto or a nonzero Steps are set.
func (p *PAP) migrationConfigured() bool {
	return p.migrateSource != "" && p.migrateDB != "" && (p.migrateAuto || p.migrateSteps != 0)
}

// Migrate runs the database migration configured via WithMigration.
func (p *PAP) Migrate() (err error) {
	if !p.migrationConfigured() {
		return nil
	}

	if p.migrateAuto {
		p.migrateSteps = 0
	}

	source := p.migrateSource
	switch {
	case strings.EqualFold(source, "*embed*"):
		err = migrate.PostgresEmbedded(migrations.PostgreSQL, p.migrateDB, p.migrateSteps, p.logger)

	default:
		if !strings.HasPrefix(source, "file://") {
			source, _ = filepath.Abs(source)
			source = "file://" + source
		}
		err = migrate.Postgres(source, p.migrateDB, p.migrateSteps, p.logger)
	}

	if err != nil && !errors.Is(err, migrate2.ErrNoChange) {
		p.logger.Error("pap database migration failed", "auto", p.migrateAuto, "steps", p.migrateSteps, "err", err)
		return
	}

	if err == nil {
		p.logger.Info("pap database migration completed successfully")
	} else {
		p.logger.Info("pap database migration completed; no changes")
		err = nil
	}
	return
}
