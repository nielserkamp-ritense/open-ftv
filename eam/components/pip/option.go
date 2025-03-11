package pip

import (
	"path/filepath"

	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip/network"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// Option represents the function signature for options when creating a new PAP.
type Option func(p *pip)

// WithFileStore adds a file storage location to the PIP.
func WithFileStore(fileStore string, recurse bool) Option {
	return func(p *pip) {
		if attrStore, _ := filepath.Abs(filepath.Join(fileStore, "attributes")); validPath(attrStore) {
			p.attrStore = attrStore
		}

		if entityStore, _ := filepath.Abs(filepath.Join(fileStore, "entities")); validPath(entityStore) {
			p.entityStore = entityStore
		}

		p.recurse = recurse
	}
}

// WithPullConfigs adds the path where pull configurations can be found.
func WithPullConfigs(path string) Option {
	return func(p *pip) {
		if pullManager, err := network.NewManager(network.ManagerParams{
			Ctx:           p.ctx,
			Path:          path,
			Logger:        p.logger,
			NewAttributes: p.newAttributes,
			Attributes:    p,
			Entities:      p,
		}); err != nil {
			p.logger.Error("failed to initialize pull manager", "path", path, "error", err)
		} else {
			p.pullManager = pullManager
		}
	}
}

// WithPersistence connects the PIP to persistent storage.
//
// By default, a PIP is created with an in-memory KV-cache.
func WithPersistence(store store.Store, basePath string) Option {
	return func(p *pip) {
		_ = p.store.Close()
		p.store = store
		p.attributePersist = NewAttributeStore(p.ctx, store, basePath)
		p.entityPersist = NewEntityStore(p.ctx, store, basePath)
	}
}

// WithFactories adds instance factories for attribute and/or entity sets.
func WithFactories(a models.AttributesBuilder, e models.EntitiesBuilder) Option {
	return func(p *pip) {
		p.newAttributes = a
		p.newEntities = e
	}
}
