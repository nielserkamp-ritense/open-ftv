// Package pip contains all logic for a functional component acting as the Policy Information Point.
package pip

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// PIP represents the interface for a Policy Information Point.
type PIP interface {
	models.AttributeSet
	models.EntitySet

	NewAttributeSet() models.AttributeSet
	NewEntitySet() models.EntitySet
	CollectAttributesFromRequest(req *components.Request) (a models.AttributeSet, newURI string)
}

// Config represents the configuration parameters for instantiating a new Policy Information Point.
//
// The Logger parameter must not be nil!
//
// If the NewAttributes parameter is nil, the default attribute set builder will be used.
// If the NewEntities parameter is nil, the default entity set builder will be used.
type Config struct {
	Ctx           context.Context
	Store         string
	Recurse       bool
	Logger        *slog.Logger
	NewAttributes models.AttributesBuilder
	NewEntities   models.EntitiesBuilder
}

// New instantiates a new Policy Information Point.
func New(cfg Config) PIP {
	var attrStore, entityStore string

	if cfg.Store != "" {
		attrStore, _ = filepath.Abs(filepath.Join(cfg.Store, "attributes"))
		entityStore, _ = filepath.Abs(filepath.Join(cfg.Store, "entities"))

		if !validPath(attrStore) {
			attrStore = ""
		}
		if !validPath(entityStore) {
			entityStore = ""
		}
	}

	if cfg.NewAttributes == nil {
		cfg.NewAttributes = models.NewAttributeSet
	}
	if cfg.NewEntities == nil {
		cfg.NewEntities = models.NewEntitySet
	}

	p := &pip{
		ctx:           cfg.Ctx,
		recurse:       cfg.Recurse,
		attrStore:     attrStore,
		entityStore:   entityStore,
		logger:        cfg.Logger,
		newAttributes: cfg.NewAttributes,
		attributes:    cfg.NewAttributes(),
		newEntities:   cfg.NewEntities,
		entities:      cfg.NewEntities(),
	}

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
	attributeWatcher *fsnotify.Watcher
	attributeTimer   *time.Timer
	attributeUpdates []string
	attributeDeletes []string
	entityWatcher    *fsnotify.Watcher
	entityTimer      *time.Timer
	entityUpdates    []string
	entityDeletes    []string
	events           models.EventSink
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
