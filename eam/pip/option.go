package pip

import (
	"os"
	"path/filepath"

	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip/network"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// Option represents the function signature for options when creating a new PAP.
type Option func(p *PIP)

// WithPullConfigs adds the path where pull configurations can be found.
func WithPullConfigs(path string) Option {
	return func(p *PIP) {
		if pullManager, err := network.NewManager(network.ManagerParams{
			Ctx:          p.ctx,
			Path:         path,
			Logger:       p.logger,
			AddAttribute: p.AddAttribute,
			GetAttribute: p.GetAttributeValue,
			AddEntity:    p.AddEntity,
			// AddRelation:  p.AddRelation,
		}); err != nil {
			p.logger.Error("failed to initialize pull manager", "path", path, "error", err)
		} else {
			p.pullManager = pullManager
		}
	}
}

// WithKeyValueDB connects the PIP to persistent storage.
//
// By default, a PIP is created with an in-memory key-value cache.
func WithKeyValueDB(store store.Store, basePath string) Option {
	return func(p *PIP) {
		if p.kvStore != nil {
			_ = p.kvStore.Close()
		}

		p.kvStore = store

		base := convert.ForceSuffix(basePath, "/")
		p.attributeDB = NewAttributeStore(store, base+"attribute/")
		p.entityDB = NewEntityStore(store, base+"entity/")
	}
}

// WithPostgresDB connects the PAP to a persistent PostgreSQL backend.
func WithPostgresDB(db *PostgresDB) Option {
	return func(p *PIP) {
		p.attributeDB = db
		p.entityDB = db
		// p.relationDB = db
	}
}

// WithFileStore adds a file storage location to the PIP.
func WithFileStore(fileStore string, recurse bool) Option {
	return func(p *PIP) {
		if fileStore != "" {
			if as, _ := filepath.Abs(filepath.Join(fileStore, "attributes")); validPath(as) {
				p.attrStore = as
			}

			if es, _ := filepath.Abs(filepath.Join(fileStore, "entities")); validPath(es) {
				p.entityStore = es
			}

			p.recurse = recurse
		}
	}
}

// WithDynamicReporter connects the PAP with an event handler for dynamic data changes.
func WithDynamicReporter(reporter ReportDynamicData) Option {
	return func(p *PIP) {
		p.dynamicData.dynReporter = reporter
	}
}

func validPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
