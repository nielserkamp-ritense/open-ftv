// Package pip contains all logic for a functional component acting as the Policy Information Point.
package pip

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip/network"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/storage/valkeyrie/memory"
)

// PIP represents the interface for a Policy Information Point.
type PIP interface {
	models.AttributeSet
	models.EntitySet

	NewAttributeSet() models.AttributeSet
	NewEntitySet() models.EntitySet
}

// New instantiates a new Policy Information Point.
//
// By default, a PIP uses an in-memory KV-cache.
// Use the WithPersistence() option to connect a PIP to persistent storage.
func New(ctx context.Context, logger *slog.Logger, options ...Option) PIP {
	if ctx == nil {
		ctx = context.Background()
	}

	s := memory.New()

	p := &pip{
		ctx:              ctx,
		logger:           logger,
		newAttributes:    models.NewAttributeSet,
		newEntities:      models.NewEntitySet,
		store:            s,
		attributePersist: NewAttributeStore(ctx, s, ""),
	}

	for i := range options {
		options[i](p)
	}

	p.attributes = p.newAttributes()
	p.entities = p.newEntities()

	p.loadFromStore()

	if p.logger.Enabled(nil, slog.LevelDebug) {
		p.logger.Debug("pip initialized", "attributeStore", p.attrStore, "entityStore", p.entityStore,
			"attributes", models.MapFromAttributes(p.attributes), "entities", p.entitiesToMap())
	} else {
		p.logger.Info("pip initialized", "attributeStore", p.attrStore, "entityStore", p.entityStore)
	}

	return p
}

// NewAttributeSet implements the PIP interface.
func (p *pip) NewAttributeSet() models.AttributeSet {
	return p.newAttributes()
}

// NewEntitySet implements the PIP interface.
func (p *pip) NewEntitySet() models.EntitySet {
	return p.newEntities()
}

// AddAttribute implements the AttributeSet interface.
//
// Use this to add a default attribute to the PIP.
func (p *pip) AddAttribute(key string, value any) {
	p.attributes.AddAttribute(key, value)
}

// AddAttributeWithType implements the AttributeSet interface.
//
// Use this to add a default attribute to the PIP.
func (p *pip) AddAttributeWithType(key string, value any, tp string) {
	p.attributes.AddAttributeWithType(key, value, tp)
}

// AddOriginalAttribute implements the AttributeSet interface.
//
// Use this to add a default attribute to the PIP.
func (p *pip) AddOriginalAttribute(key string, value, original any, tp string) {
	p.attributes.AddOriginalAttribute(key, value, original, tp)
}

// GetAttribute implements the AttributeSet interface.
//
// Use this to read a default attribute from the PIP.
func (p *pip) GetAttribute(key string) models.Attribute {
	return p.attributes.GetAttribute(key)
}

// GetAttributeValue implements the AttributeSet interface.
//
// Use this to read a default attribute value from the PIP.
func (p *pip) GetAttributeValue(key string) any {
	return p.attributes.GetAttributeValue(key)
}

// RemoveAttribute implements the AttributeSet interface.
//
// Use this to remove a default attribute from the PIP.
func (p *pip) RemoveAttribute(key string) {
	p.attributes.RemoveAttribute(key)
}

// IterateAttributes implements the AttributeSet interface.
//
// Use this to iterate through all default attributes from the PIP.
func (p *pip) IterateAttributes(f models.AttributeIterator) {
	p.attributes.IterateAttributes(f)
}

// MergeAttributes implements the AttributeSet interface.
//
// Use this to merge an attribute set into the default attributes of the PIP.
func (p *pip) MergeAttributes(in ...models.AttributeSet) {
	p.attributes.MergeAttributes(in...)
}

// AddEntity implements the EntitySet interface.
//
// Use this to add an entity to the PIP.
func (p *pip) AddEntity(entity models.Entity) {
	p.entities.AddEntity(entity)
}

// GetEntity implements the EntitySet interface.
//
// Use this to read an entity from the PIP.
func (p *pip) GetEntity(uid string) models.Entity {
	return p.entities.GetEntity(uid)
}

// RemoveEntity implements the EntitySet interface.
//
// Use this to remove an entity from the PIP.
func (p *pip) RemoveEntity(uid string) {
	p.entities.RemoveEntity(uid)
}

// IterateEntities implements the EntitySet interface.
//
// Use this to iterate through all entities from the PIP.
func (p *pip) IterateEntities(f models.EntityIterator) {
	p.entities.IterateEntities(f)
}

// MergeEntities implements the EntitySet interface.
//
// Use this to merge an attribute set into the entities of the PIP.
func (p *pip) MergeEntities(in ...models.EntitySet) {
	p.entities.MergeEntities(in...)
}

// MarshalJSON implements the json.Marshaller interface.
func (p *pip) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

type pip struct {
	recurse          bool
	attrStore        string
	entityStore      string
	logger           *slog.Logger
	ctx              context.Context
	newAttributes    models.AttributesBuilder
	attributes       models.AttributeSet
	newEntities      models.EntitiesBuilder
	entities         models.EntitySet
	pullManager      network.Manager
	attributeWatcher *fsnotify.Watcher
	attributeTimer   *time.Timer
	attributeUpdates []string
	attributeDeletes []string
	entityWatcher    *fsnotify.Watcher
	entityTimer      *time.Timer
	entityUpdates    []string
	entityDeletes    []string
	events           models.EventSink
	store            store.Store
	attributePersist AttributePersistence
	mutex            sync.RWMutex
}

func validPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (p *pip) entitiesToMap() map[string]any {
	out := make(map[string]any)

	p.entities.IterateEntities(func(entity models.Entity) {
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
