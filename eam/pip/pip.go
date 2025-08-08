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

// PIP represents the interface for a Policy Information Point.
type PIP struct {
	recurse          bool
	attrStore        string
	entityStore      string
	logger           *slog.Logger
	ctx              context.Context
	pullManager      network.Manager
	attributeWatcher *fsnotify.Watcher
	attributeTimer   *time.Timer
	attributeUpdates []string
	attributeDeletes []string
	entityWatcher    *fsnotify.Watcher
	entityTimer      *time.Timer
	entityUpdates    []string
	entityDeletes    []string
	eventSinks       []models.EventSink
	store            store.Store
	attributePersist AttributePersistence
	entityPersist    EntityPersistence
	eventMutex       sync.RWMutex
	bundleMutex      sync.RWMutex
}

// New instantiates a new Policy Information Point.
//
// By default, a PIP uses an in-memory key-value cache.
// Use the WithPersistence() option to connect a PIP to persistent storage.
func New(ctx context.Context, logger *slog.Logger, options ...Option) *PIP {
	if ctx == nil {
		ctx = context.Background()
	}

	s := memory.New()
	ap := NewAttributeStore(ctx, s, "attribute")
	ep := NewEntityStore(ctx, s, "entity")

	p := &PIP{
		ctx:              ctx,
		logger:           logger,
		store:            s,
		attributePersist: ap,
		entityPersist:    ep,
	}

	for i := range options {
		options[i](p)
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

		if p.entityPersist != ep || p.attributePersist != ap {
			args = append(args, "persistence", true)
		}

		if p.logger.Enabled(nil, slog.LevelDebug) {
			attrs, _ := p.attributePersist.List()
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

func (p *PIP) sendEvent(eventType models.EventType, key string) {
	p.eventMutex.Lock()
	for i := range p.eventSinks {
		p.eventSinks[i].Handle(eventType, key)
	}
	defer p.eventMutex.Unlock()
}
