// Package pip contains all logic for a functional component acting as the Policy Information Point.
package pip

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip/network"
)

// ErrNoPersistence is returned by New when no persistence backend was configured via
// WithKeyValueDB() or WithPostgresDB().
var ErrNoPersistence = errors.New("pip: no persistence backend configured")

// logArgsCapacity is the pre-allocated size of the slog argument slice built in logInitialized.
const logArgsCapacity = 8

// ReportDynamicData is the function signature for reporting changes in runtime PIP data.
type ReportDynamicData func(data map[string]any)

// PIP represents the interface for caching and retrieving attribute-, entity- and/or relation-data.
type PIP struct {
	recurse          bool
	attrStore        string
	entityStore      string
	logger           *slog.Logger
	ctx              context.Context
	pullManager      network.Manager
	attributeDB      AttributePersister
	attributeWatcher *fsnotify.Watcher
	attributeTimer   *time.Timer
	attributeUpdates []string
	attributeDeletes []string
	entityDB         EntityPersister
	entityWatcher    *fsnotify.Watcher
	entityTimer      *time.Timer
	entityUpdates    []string
	entityDeletes    []string
	relationDB       RelationPersister
	kvStore          store.Store
	eventSinks       []models.EventSink
	eventMutex       sync.RWMutex
	bundleMutex      sync.RWMutex
	dynamicData      dynamicData
}

// New instantiates a new Policy Information Point.
//
// A PIP requires a persistence backend: use WithKeyValueDB() or WithPostgresDB() to connect one. Without
// one, New returns ErrNoPersistence. WithFileStore() additionally loads attributes and entities from local
// files into whichever backend is configured.
func New(ctx context.Context, logger *slog.Logger, options ...Option) (_ *PIP, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	w := newFileWatcher()

	p := &PIP{
		ctx:              ctx,
		logger:           logger,
		attributeUpdates: make([]string, 0),
		attributeDeletes: make([]string, 0),
		attributeWatcher: w,
		entityUpdates:    make([]string, 0),
		entityDeletes:    make([]string, 0),
		entityWatcher:    w,
		dynamicData: dynamicData{
			attributes: models.NewAttributeSet(),
			entities:   models.NewEntitySet(),
			relations:  models.NewRelationSet(nil),
		},
	}

	// watchFiles() takes ownership of w and closes it once started; if we fail before that, closing it here.
	defer func() {
		if err != nil {
			p.cleanup(w)
		}
	}()

	for i := range options {
		options[i](p)
	}

	if p.attributeDB == nil || p.entityDB == nil /* || p.relationDB == nil */ {
		logger.Error("pip: no persistence backend configured")
		return nil, ErrNoPersistence
	}

	p.loadFromStore()
	p.logInitialized()

	return p, nil
}

// newFileWatcher creates a filesystem watcher, or returns nil if file handles are exhausted.
func newFileWatcher() *fsnotify.Watcher {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil
	}

	return w
}

// cleanup releases resources acquired before New() failed.
func (p *PIP) cleanup(w *fsnotify.Watcher) {
	if w != nil {
		_ = w.Close()
	}

	if p.pullManager != nil {
		p.pullManager.Stop()
	}
}

func (p *PIP) hasFileStore() bool {
	return (p.attrStore != "" && p.attrStore != "/") || (p.entityStore != "" && p.entityStore != "/")
}

func (p *PIP) logInitialized() {
	if !p.logger.Enabled(p.ctx, slog.LevelInfo) {
		return
	}

	args := make([]any, 0, logArgsCapacity)

	if p.hasFileStore() {
		if p.attrStore != "" {
			args = append(args, "attributeStore", p.attrStore)
		}

		if p.entityStore != "" {
			args = append(args, "entityStore", p.entityStore)
		}

		args = append(args, "recurse", p.recurse)
	}

	if p.kvStore == nil {
		args = append(args, "persistence", true)
	}

	if p.logger.Enabled(p.ctx, slog.LevelDebug) {
		attrs, _ := p.attributeDB.ListAttributes(p.ctx)
		args = append(args, "attributes", attrs, "entities", p.entitiesToMap())
		p.logger.Debug("pip initialized", args...)

		return
	}

	p.logger.Info("pip initialized", args...)
}

// MarshalJSON implements the json.Marshaler interface.
func (p *PIP) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

func (p *PIP) entitiesToMap() map[string]any {
	out := make(map[string]any)

	p.IterateEntities(func(entity *models.Entity) {
		out[entity.UID()] = struct {
			UID        string         `json:"UID,omitempty"`
			Attributes map[string]any `json:"attributes,omitempty"`
			Parents    []string       `json:"parents,omitempty"`
		}{
			UID:        entity.UID(),
			Attributes: models.MapFromAttributes(entity.Attributes()),
			Parents:    entity.Parents(),
		}
	})

	return out
}

// AddEventSink adds an event processor to the PIP.
func (p *PIP) AddEventSink(events models.EventSink) {
	p.eventMutex.Lock()
	p.eventSinks = append(p.eventSinks, events)
	p.eventMutex.Unlock()
}

// ReportDynamicData reports the set of dynamically added/modified attributes, entities and relations.
func (p *PIP) ReportDynamicData(f func()) {
	p.dynamicData.report(f)
}

func (p *PIP) sendEvent(eventType models.EventType, key string) {
	p.eventMutex.Lock()
	for i := range p.eventSinks {
		p.eventSinks[i].Handle(eventType, key)
	}
	defer p.eventMutex.Unlock()
}
