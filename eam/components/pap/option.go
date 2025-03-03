package pap

import (
	"github.com/kvtools/valkeyrie/store"
)

// Option represents the function signature for options when creating a new PAP.
type Option func(p *pap)

// WithLanguage sets the default policy language fopr the PAP.
func WithLanguage(language string) Option {
	return func(p *pap) {
		p.language = language
	}
}

// WithPersistence connects the PAP to persistent storage.
//
// By default, a PAP is created with an in-memory KV-cache.
func WithPersistence(store store.Store, basePath string) Option {
	return func(p *pap) {
		_ = p.store.Close()
		p.store = store
		p.persist = NewStore(p.ctx, store, basePath)
	}
}

// WithFileStore adds a file storage location to the controller.
func WithFileStore(fileStore string, recurse bool) Option {
	return func(p *pap) {
		p.fileStore = fileStore
		p.recurse = recurse
	}
}
