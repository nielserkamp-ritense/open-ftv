package cedar

import (
	"fmt"
	"log/slog"
	"maps"
	"strings"
	"sync"

	"github.com/cedar-policy/cedar-go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
)

// WrappedEntity is a wrapper around a cedar entity.
//
// It is read-only by design.
type WrappedEntity struct {
	uid    string
	ce     *cedar.Entity
	logger *slog.Logger
}

// NewWrappedEntity instantiates a new wrapper around a cedar entity.
func NewWrappedEntity(ce *cedar.Entity, logger *slog.Logger) models.Entity {
	return &WrappedEntity{
		uid:    cedarToUID(ce.UID),
		ce:     ce,
		logger: logger,
	}
}

// UID implements the Entity interface.
func (e *WrappedEntity) UID() string {
	return e.uid
}

// Type implements the Entity interface.
func (e *WrappedEntity) Type() string {
	return string(e.ce.UID.Type)
}

// ID implements the Entity interface.
func (e *WrappedEntity) ID() string {
	return string(e.ce.UID.ID)
}

// Attributes implements the Entity interface.
func (e *WrappedEntity) Attributes() models.AttributeSet {
	return &attributes{logger: e.logger, set: e.ce.Attributes.Map()}
}

// Parents implements the Entity interface.
func (e *WrappedEntity) Parents() []string {
	out := make([]string, 0, e.ce.Parents.Len())
	e.ce.Parents.Iterate(func(ce cedar.EntityUID) bool {
		out = append(out, cedarToUID(ce))
		return true
	})
	return out
}

// MarshalJSON implements the JSON marshaller interface.
func (e *WrappedEntity) MarshalJSON() ([]byte, error) {
	return models.NewEntity(e.Type(), e.ID(), e.Attributes(), e.Parents()...).MarshalJSON()
}

// NewEntityBuilder returns the function prototype for building a new Cedar based entity set.
func NewEntityBuilder(logger *slog.Logger) models.EntitiesBuilder {
	return func(in ...any) models.EntitySet {
		return NewEntitySet(logger, in...)
	}
}

// NewEntitySet instantiates a new Cedar based entity set.
func NewEntitySet(logger *slog.Logger, in ...any) models.EntitySet {
	a := &entities{logger: logger, set: make(cedar.EntityMap)}
	for _, p := range in {
		switch t := p.(type) {
		case models.Entity:
			a.AddEntity(t)
		case models.EntitySet:
			a.MergeEntities(t)
		}
	}
	return a
}

// AddEntity implements the EntitySet interface.
func (e *entities) AddEntity(entity models.Entity) {
	e.mutex.Lock()
	if wrapped, ok := entity.(*WrappedEntity); ok {
		e.set[wrapped.ce.UID] = *wrapped.ce
	} else {
		ce := entityToCedar(entity)
		e.set[ce.UID] = *ce
	}
	e.mutex.Unlock()
}

// GetEntity implements the EntitySet interface.
func (e *entities) GetEntity(uid string) models.Entity {
	e.mutex.RLock()
	ce, ok := e.set[uidToCedar(uid)]
	e.mutex.RUnlock()
	if !ok {
		return nil
	}
	return e.cedarToEntity(&ce)
}

// RemoveEntity implements the EntitySet interface.
func (e *entities) RemoveEntity(uid string) {
	e.mutex.Lock()
	delete(e.set, uidToCedar(uid))
	e.mutex.Unlock()
}

// IterateEntities implements the EntitySet interface.
func (e *entities) IterateEntities(f models.EntityIterator) {
	e.mutex.RLock()
	for uid := range e.set {
		ce := e.set[uid]
		f(NewWrappedEntity(&ce, e.logger))
	}
	e.mutex.RUnlock()
}

// MergeEntities implements the EntitySet interface.
func (e *entities) MergeEntities(in ...models.EntitySet) {
	e.mutex.Lock()
	for i := range in {
		if set, ok := in[i].(*entities); ok {
			set.mutex.RLock()
			maps.Copy(e.set, set.set)
			set.mutex.RUnlock()
		} else {
			in[i].IterateEntities(func(entity models.Entity) {
				ce := entityToCedar(entity)
				e.set[ce.UID] = *ce
			})
		}
	}
	e.mutex.Unlock()
}

func (e *entities) cedarToEntity(in *cedar.Entity) models.Entity {
	if in == nil || in == empty {
		return nil
	}
	return NewWrappedEntity(in, e.logger)
}

func uidToCedar(uid string) cedar.EntityUID {
	parts := strings.Split(uid, "::")
	return cedar.NewEntityUID(cedar.EntityType(parts[0]), cedar.String(parts[1]))
}

func cedarToUID(ce cedar.EntityUID) string {
	return fmt.Sprintf("%s::%s", ce.Type, ce.ID)
}

func entityToCedar(in models.Entity) *cedar.Entity {
	if in == nil {
		return nil
	}

	if e, ok := in.(*WrappedEntity); ok {
		return e.ce
	}

	parents := make([]cedar.EntityUID, len(in.Parents()))
	for i := range in.Parents() {
		parents[i] = uidToCedar(in.Parents()[i])
	}

	attrs := attributes{set: make(cedar.RecordMap)}
	attrs.MergeAttributes(in.Attributes())

	return &cedar.Entity{
		UID:        uidToCedar(in.UID()),
		Parents:    cedar.NewEntityUIDSet(parents...),
		Attributes: cedar.NewRecord(attrs.set),
	}
}

type entities struct {
	logger *slog.Logger
	set    cedar.EntityMap
	mutex  sync.RWMutex
}

var empty = &cedar.Entity{}
