// Package pip contains all logic for a functional component acting as the Policy Information Point.
package pip

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip/network"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

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
// By default, a PIP uses an in-memory key-value cache.
// Use the WithKeyValueDB or WithPostgresDB option to connect a PIP to persistent storage.
func New(ctx context.Context, logger *slog.Logger, options ...Option) *PIP {
	if ctx == nil {
		ctx = context.Background()
	}

	p := &PIP{
		ctx:              ctx,
		logger:           logger,
		attributeUpdates: make([]string, 0),
		attributeDeletes: make([]string, 0),
		entityUpdates:    make([]string, 0),
		entityDeletes:    make([]string, 0),
		dynamicData: dynamicData{
			attributes: models.NewAttributeSet(),
			entities:   models.NewEntitySet(),
			relations:  models.NewRelationSet(nil),
		},
	}

	for i := range options {
		options[i](p)
	}

	if p.attributeDB == nil || p.entityDB == nil /* || p.relationDB == nil */ {
		p.kvStore = memory.New()

		if p.attributeDB == nil {
			p.attributeDB = NewAttributeStore(p.kvStore, "attribute")
		}
		if p.entityDB == nil {
			p.entityDB = NewEntityStore(p.kvStore, "entity")
		}
		// if p.relationDB == nil {
		// 	p.relationDB = NewRelationStore(p.kvStore, "entity")
		// }
	}

	p.loadFromStore()

	if p.logger.Enabled(nil, slog.LevelInfo) {
		args := make([]any, 0, 8)

		if (p.attrStore != "" && p.attrStore != "/") || (p.entityStore != "" && p.entityStore != "/") {
			if p.attrStore != "" {
				args = append(args, "attributeStore", p.attrStore)
			}
			if p.entityStore != "" {
				args = append(args, "entityStore", p.entityStore)
			}
			args = append(args, "recurse", p.recurse)
		}

		if p.kvStore == nil && (p.entityDB != nil || p.attributeDB != nil) {
			args = append(args, "persistence", true)
		}

		if p.logger.Enabled(nil, slog.LevelDebug) {
			attrs, _ := p.attributeDB.ListAttributes(p.ctx)
			args = append(args, "attributes", attrs, "entities", p.entitiesToMap())
			p.logger.Debug("pip initialized", args...)
		} else {
			p.logger.Info("pip initialized", args...)
		}
	}
	return p
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
