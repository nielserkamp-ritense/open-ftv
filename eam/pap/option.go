package pap

import (
	"os"
	"path/filepath"

	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

// Option represents the function signature for options when creating a new PAP.
type Option func(p *PAP)

// WithLanguage sets the default policy language fopr the PAP.
func WithLanguage(language string) Option {
	return func(p *PAP) {
		p.language = language
		p.languageType = models.LanguageFromString(language)
	}
}

// WithKeyValueDB connects the PAP to a persistent KV backend.
//
// By default, a PAP is created with an in-memory key-value cache.
func WithKeyValueDB(store store.Store, basePath string) Option {
	return func(p *PAP) {
		if p.kvStore != nil {
			_ = p.kvStore.Close()
		}

		p.kvStore = store
		p.policyDB = NewKeyValueDB(store, basePath)
		p.bundleDB = bundles.NewPersistence(p.ctx, store, basePath)
	}
}

// WithPgPool connects the PAP to a PostgreSQL connection pool.
func WithPgPool(pool *postgresql.Postgres) Option {
	return func(p *PAP) {
		p.languageDB = NewLanguageDBWithPool(pool)
		p.tagDB = NewTagDBWithPool(pool)
		p.policyDB = NewPolicyDBWithPool(pool)
	}
}

// WithPolicyDB connects the PAP to a PostgreSQL backend for managing policies.
func WithPolicyDB(db *PolicyDB) Option {
	return func(p *PAP) {
		p.policyDB = db
	}
}

// WithMigration initializes database migrations.
//
// *source* defines the location of the migration scripts.
// *db* is the URL used to open the database.
//
// If *auto* is set to true, the migration handler will upgrade the database to the highest level.
// Otherwise, it will use the given *steps* to determine if it needs to migrate the database up (positive number) or down (negative number).
func WithMigration(source, db string, auto bool, steps int) Option {
	return func(p *PAP) {
		p.migrateSource = source
		p.migrateDB = db
		p.migrateAuto = auto
		p.migrateSteps = steps
	}
}

// WithFileStore adds a file storage location to the controller.
func WithFileStore(fileStore string, recurse bool) Option {
	return func(p *PAP) {
		if fileStore != "" {
			if ps, _ := filepath.Abs(fileStore); validPath(ps) {
				p.policyStore = ps
				p.recurse = recurse
			}
		}
	}
}

func validPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
