package pap

import (
	"os"
	"path/filepath"

	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
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

// WithPersistence connects the PAP to persistent storage.
//
// By default, a PAP is created with an in-memory key-value cache.
func WithPersistence(store store.Store, basePath string) Option {
	return func(p *PAP) {
		if p.store != nil {
			_ = p.store.Close()
		}

		p.store = store
		p.persist = NewPersistence(p.ctx, store, basePath)
		p.deployer = bundles.NewPersistence(p.ctx, store, basePath)
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
