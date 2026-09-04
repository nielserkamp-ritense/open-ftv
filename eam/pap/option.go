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
func WithKeyValueDB(store store.Store, basePath string) Option {
	return func(p *PAP) {
		if p.kvStore != nil {
			_ = p.kvStore.Close()
		}

		p.kvStore = store
		p.policyDB = NewKeyValueDB(store, basePath)
		p.bundleDB = bundles.NewKeyValueDB(store, basePath)
	}
}

// WithPgPool connects the PAP to a PostgreSQL connection pool.
func WithPgPool(pool *postgresql.Postgres) Option {
	return func(p *PAP) {
		p.languageDB = NewLanguageDBWithPool(pool)
		p.tagDB = NewTagDBWithPool(pool)
		p.policyDB = NewPolicyDBWithPool(pool)
		p.bundleDB = bundles.NewPostgresDBWithPool(pool)
	}
}

// WithMigration sets the URL of the database that Migrate() brings up to the latest level with the
// migration scripts embedded in the binary.
func WithMigration(db string) Option {
	return func(p *PAP) {
		p.migrateDB = db
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
