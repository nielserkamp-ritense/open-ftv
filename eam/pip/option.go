package pip

import (
	"os"
	"path/filepath"

	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip/network"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/warc"
)

// Option represents the function signature for options when creating a new PAP.
type Option func(p *pip)

// WithFileStore adds a file storage location to the PIP.
func WithFileStore(fileStore string, recurse bool) Option {
	return func(p *pip) {
		if as, _ := filepath.Abs(filepath.Join(fileStore, "attributes")); validPath(as) {
			p.attrStore = as
		}

		if es, _ := filepath.Abs(filepath.Join(fileStore, "entities")); validPath(es) {
			p.entityStore = es
		}

		p.recurse = recurse
	}
}

// WithPullConfigs adds the path where pull configurations can be found.
//
// When the PIP also has a WARC log configured (see WithWARC), make sure that option
// precedes this one; every pull request/response pair is then logged to the WARC file
// and each decoded attribute/entity is registered with a "Logged" source reference.
func WithPullConfigs(path string) Option {
	return func(p *pip) {
		if pullManager, err := network.NewManager(network.ManagerParams{
			Ctx:           p.ctx,
			Path:          path,
			Logger:        p.logger,
			NewAttributes: p.newAttributes,
			Attributes:    p,
			Entities:      p,
			WARC:          p.warc,
			Recorder: func(kind, key, traceID, spanID, warcFile, version string) {
				p.RecordSourceRef(SourceRef{Kind: kind, Key: key, TraceID: traceID, SpanID: spanID, WARCFile: warcFile, Version: version})
			},
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

		base := convert.ForceSuffix(basePath, "/")
		p.attributePersist = NewAttributeStore(p.ctx, store, base+"attribute/")
		p.entityPersist = NewEntityStore(p.ctx, store, base+"entity/")
		p.sourceRefs = newSourceRefStore(p.ctx, store, base+"sourceref/")
	}
}

// WithEventSink connects an event sink to the PIP.
//
// The PIP emits an event for every added, replaced or removed attribute and entity.
func WithEventSink(sink models.EventSink) Option {
	return func(p *pip) {
		p.events = sink
	}
}

// WithWARC connects a WARC log to the PIP.
//
// All external information exchanges (outgoing pull requests and their responses) are
// appended to the WARC log so that an Authorization Decision Log can reference them as
// "Logged" source references (indexed on trace_id + span_id).
//
// This option should precede WithPullConfigs.
func WithWARC(w *warc.Writer) Option {
	return func(p *pip) {
		p.warc = w
	}
}

// WithFactories adds instance factories for attribute and/or entity sets.
func WithFactories(a models.AttributesBuilder, e models.EntitiesBuilder) Option {
	return func(p *pip) {
		p.newAttributes = a
		p.newEntities = e
	}
}

func validPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
