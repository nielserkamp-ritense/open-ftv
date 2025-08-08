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

// WithPersistence connects the PIP to persistent storage.
//
// By default, a PIP is created with an in-memory key-value cache.
func WithPersistence(store store.Store, basePath string) Option {
	return func(p *PIP) {
		_ = p.store.Close()
		p.store = store

		base := convert.ForceSuffix(basePath, "/")
		p.attributePersist = NewAttributeStore(p.ctx, store, base+"attribute/")
		p.entityPersist = NewEntityStore(p.ctx, store, base+"entity/")
	}
}

func validPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
