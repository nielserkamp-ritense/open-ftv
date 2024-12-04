// Package pip contains all logic for a functional component acting as the Policy Information Point.
package pip

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/standards"
)

// PIP represents the interface for a Policy Information Point.
type PIP interface {
	standards.AttributeSet
	standards.EntitySet
	CollectAttributesFromRequest(req *standards.Request) (a standards.AttributeSet, newURI string)
}

// New instantiates a new Policy Information Point.
//
// The logger parameter must not be nil!
//
// If the newAttributes parameter is nil, the default attribute set builder will be used.
// If the newEntities parameter is nil, the default entity set builder will be used.
func New(store string, recurse bool, logger *slog.Logger, newAttributes standards.AttributesBuilder, newEntities standards.EntitiesBuilder) PIP {
	var attrStore, entityStore string

	if store != "" {
		attrStore, _ = filepath.Abs(filepath.Join(store, "attributes"))
		entityStore, _ = filepath.Abs(filepath.Join(store, "entities"))

		if !validPath(attrStore) {
			attrStore = ""
		}
		if !validPath(entityStore) {
			entityStore = ""
		}
	}

	if newAttributes == nil {
		newAttributes = standards.NewAttributeSet
	}
	if newEntities == nil {
		newEntities = standards.NewEntitySet
	}

	p := &pip{
		recurse:       recurse,
		attrStore:     attrStore,
		entityStore:   entityStore,
		logger:        logger,
		newAttributes: newAttributes,
		attributes:    newAttributes(),
		newEntities:   newEntities,
		entities:      newEntities(),
	}

	p.load()

	if p.logger.Enabled(nil, slog.LevelDebug) {
		p.logger.Debug("pip initialized", "attributeStore", p.attrStore, "entityStore", p.entityStore,
			"attributes", standards.MapFromAttributes(p.attributes), "entities", p.entitiesToMap())
	} else {
		p.logger.Info("pip initialized", "attributeStore", p.attrStore, "entityStore", p.entityStore)
	}

	return p
}

func (p *pip) entitiesToMap() map[string]any {
	out := make(map[string]any)

	p.entities.IterateEntities(func(entity standards.Entity) {
		out[entity.UID()] = struct {
			UID        string         `json:"UID,omitempty"`
			Attributes map[string]any `json:"attributes,omitempty"`
			Parents    []string       `json:"parents,omitempty"`
		}{
			UID:        entity.UID(),
			Attributes: standards.MapFromAttributes(entity.Attributes()),
			Parents:    entity.Parents(),
		}
	})

	return out
}

// CollectAttributesFromRequest uses the given PBAC authorization request and other inputs
// to collect a set of attributes to be used by the Policy Decision Point.
//
// The default attributes stored in the PIP will be collected first.
// AttributeSet from the request will overwrite default attributes when the keys are equal.
func (p *pip) CollectAttributesFromRequest(req *standards.Request) (standards.AttributeSet, string) {
	a := p.newAttributes(p.attributes)
	a.AddAttribute(standards.AttrRequestTime, time.Now().UTC())

	newURI := p.testHeaders(req, a)
	if newURI == "" && req.Resource != nil {
		newURI = req.Resource.ID()
	}

	p.determineURL(req, a)
	p.decodeBody(req, a)

	if req.Principal != nil {
		a.AddAttribute(standards.AttrPrincipal, req.Principal.UID())
	}
	if req.Action != nil {
		a.AddAttribute(standards.AttrAction, req.Action.UID())
	}
	if req.Resource != nil {
		a.AddAttribute(standards.AttrResource, req.Resource.UID())
	}

	if len(req.Attributes) > 0 {
		for k := range req.Attributes {
			a.AddAttribute(k, req.Attributes[k])
		}
	}

	if p.logger.Enabled(nil, slog.LevelDebug) {
		kv := make(map[string]any)
		a.IterateAttributes(func(k string, v any) {
			kv[k] = v
		})
		p.logger.Debug("attributes collected", "request-uid", req.UID, "attributes", kv)
	}

	return a, newURI
}

// AddAttribute implements the AttributeSet interface.
//
// Use this to add a default attribute to the PIP.
func (p *pip) AddAttribute(key string, value any) {
	p.attributes.AddAttribute(key, value)
}

// GetAttribute implements the AttributeSet interface.
//
// Use this to read a default attribute from the PIP.
func (p *pip) GetAttribute(key string) any {
	return p.attributes.GetAttribute(key)
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
func (p *pip) IterateAttributes(f standards.AttributeIterator) {
	p.attributes.IterateAttributes(f)
}

// MergeAttributes implements the AttributeSet interface.
//
// Use this to merge an attribute set into the default attributes of the PIP.
func (p *pip) MergeAttributes(in ...standards.AttributeSet) {
	p.attributes.MergeAttributes(in...)
}

// AddEntity implements the EntitySet interface.
//
// Use this to add an entity to the PIP.
func (p *pip) AddEntity(entity standards.Entity) {
	p.entities.AddEntity(entity)
}

// GetEntity implements the EntitySet interface.
//
// Use this to read an entity from the PIP.
func (p *pip) GetEntity(uid string) standards.Entity {
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
func (p *pip) IterateEntities(f standards.EntityIterator) {
	p.entities.IterateEntities(f)
}

// MergeEntities implements the EntitySet interface.
//
// Use this to merge an attribute set into the entities of the PIP.
func (p *pip) MergeEntities(in ...standards.EntitySet) {
	p.entities.MergeEntities(in...)
}

type pip struct {
	recurse       bool
	attrStore     string
	entityStore   string
	logger        *slog.Logger
	newAttributes standards.AttributesBuilder
	attributes    standards.AttributeSet
	newEntities   standards.EntitiesBuilder
	entities      standards.EntitySet
}

func validPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
