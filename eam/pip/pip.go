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
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/warc"
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
	ap := NewAttributeStore(ctx, s, "attribute")
	ep := NewEntityStore(ctx, s, "entity")

	p := &pip{
		ctx:              ctx,
		logger:           logger,
		newAttributes:    models.NewAttributeSet,
		newEntities:      models.NewEntitySet,
		store:            s,
		attributePersist: ap,
		entityPersist:    ep,
		sourceRefs:       newSourceRefStore(ctx, s, "sourceref"),
	}

	for i := range options {
		options[i](p)
	}

	p.loadFromStore()

	if p.logger.Enabled(nil, slog.LevelInfo) {
		args := make([]any, 0, 8)

		if p.attrStore != "" || p.entityStore != "" {
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
	_ = p.addAttribute(models.NewAttribute(key, value))
}

// AddAttributeWithType implements the AttributeSet interface.
//
// Use this to add a default attribute to the PIP.
func (p *pip) AddAttributeWithType(key string, value any, tp string) {
	_ = p.addAttribute(models.NewAttributeWithType(key, value, tp))
}

// AddOriginalAttribute implements the AttributeSet interface.
//
// Use this to add a default attribute to the PIP.
func (p *pip) AddOriginalAttribute(key string, value, original any, tp string) {
	_ = p.addAttribute(models.NewOriginalAttribute(key, value, original, tp))
}

func (p *pip) addAttribute(a models.Attribute) error {
	prev, ix, err := p.attributePersist.Read(a.Key())
	if err != nil || prev == nil {
		if _, err = p.attributePersist.Create(a); err == nil {
			p.emit(models.AttributeAdded, a.Key())
		}
	} else {
		if _, err = p.attributePersist.Update(prev, ix, a); err == nil {
			p.emit(models.AttributeReplaced, a.Key())
		}
	}
	return err
}

// GetAttribute implements the AttributeSet interface.
//
// Use this to read a default attribute from the PIP.
func (p *pip) GetAttribute(key string) models.Attribute {
	a, _, _ := p.attributePersist.Read(key)
	return a
}

// GetAttributeValue implements the AttributeSet interface.
//
// Use this to read a default attribute value from the PIP.
func (p *pip) GetAttributeValue(key string) any {
	if a, _, _ := p.attributePersist.Read(key); a != nil {
		return a.Value()
	}
	return nil
}

// RemoveAttribute implements the AttributeSet interface.
//
// Use this to remove a default attribute from the PIP.
func (p *pip) RemoveAttribute(key string) {
	if prev, ix, err := p.attributePersist.Read(key); err == nil {
		if _, err = p.attributePersist.Delete(prev, ix); err == nil {
			p.emit(models.AttributeRemoved, key)
		}
	}
}

// IterateAttributes implements the AttributeSet interface.
//
// Use this to iterate through all default attributes from the PIP.
func (p *pip) IterateAttributes(f models.AttributeIterator) {
	if list, err := p.attributePersist.List(); err == nil {
		for i := range list {
			f(list[i])
		}
	}
}

// MergeAttributes implements the AttributeSet interface.
//
// Use this to merge an attribute set into the default attributes of the PIP.
func (p *pip) MergeAttributes(in ...models.AttributeSet) {
	for i := range in {
		in[i].IterateAttributes(func(attr models.Attribute) {
			_ = p.addAttribute(attr)
		})
	}
}

// AddEntity implements the EntitySet interface.
//
// Use this to add an entity to the PIP.
func (p *pip) AddEntity(entity models.Entity) {
	prev, ix, err := p.entityPersist.Read(entity.UID())
	if err != nil || prev == nil {
		if _, err = p.entityPersist.Create(entity); err == nil {
			p.emit(models.EntityAdded, entity.UID())
		}
	} else {
		if _, err = p.entityPersist.Update(prev, ix, entity); err == nil {
			p.emit(models.EntityReplaced, entity.UID())
		}
	}
}

// GetEntity implements the EntitySet interface.
//
// Use this to read an entity from the PIP.
func (p *pip) GetEntity(uid string) models.Entity {
	e, _, _ := p.entityPersist.Read(uid)
	return e
}

// RemoveEntity implements the EntitySet interface.
//
// Use this to remove an entity from the PIP.
func (p *pip) RemoveEntity(uid string) {
	if prev, ix, err := p.entityPersist.Read(uid); err == nil {
		if _, err = p.entityPersist.Delete(prev, ix); err == nil {
			p.emit(models.EntityRemoved, uid)
		}
	}
}

// IterateEntities implements the EntitySet interface.
//
// Use this to iterate through all entities from the PIP.
func (p *pip) IterateEntities(f models.EntityIterator) {
	if list, err := p.entityPersist.List(); err == nil {
		for i := range list {
			f(list[i])
		}
	}
}

// MergeEntities implements the EntitySet interface.
//
// Use this to merge an attribute set into the entities of the PIP.
func (p *pip) MergeEntities(in ...models.EntitySet) {
	for i := range in {
		in[i].IterateEntities(func(e models.Entity) {
			p.AddEntity(e)
		})
	}
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
	newEntities      models.EntitiesBuilder
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
	warc             *warc.Writer
	sourceRefs       *sourceRefStore
	store            store.Store
	attributePersist AttributePersistence
	entityPersist    EntityPersistence
	mutex            sync.RWMutex
}

// emit sends an event to the configured EventSink, if any.
func (p *pip) emit(t models.EventType, key string) {
	if p.events != nil {
		p.events.Handle(t, key)
	}
}

func (p *pip) entitiesToMap() map[string]any {
	out := make(map[string]any)

	p.IterateEntities(func(entity models.Entity) {
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
