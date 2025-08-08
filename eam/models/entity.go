package models

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

// EntityUID formats the unique identifier (UID) for an entity.
func EntityUID(ns, id string) string {
	return fmt.Sprintf("%s::%s", ns, id)
}

// Entity contains the details of an entity.
//
// Entity is an immutable object and is by design safe for use in concurrent go-routines.
type Entity struct {
	uid         string              // unique identifier (ns::id).
	ns          string              // name-space.
	id          string              // unique identifier within the name-space.
	title       string              // short description.
	description string              // long description.
	attrs       *AttributeSet       // optional attributes.
	parents     []string            // optional set of unique identifiers of parent entities.
	tags        map[string]struct{} // tags associated with the entity.
}

// NewEntity instanties a new standard entity.
func NewEntity(ns, id string, attrs *AttributeSet, parents ...string) *Entity {
	if attrs == nil {
		attrs = NewAttributeSet()
	}

	return &Entity{
		uid:     EntityUID(ns, id),
		ns:      ns,
		id:      id,
		attrs:   attrs,
		parents: parents,
		tags:    make(map[string]struct{}),
	}
}

// WithTitle adds an optional title to the Entity.
func (e *Entity) WithTitle(title string) *Entity {
	e.title = title
	return e
}

// WithDescription adds an optional description to the Entity.
func (e *Entity) WithDescription(desc string) *Entity {
	e.description = desc
	return e
}

// WithTags annotates the Entity with the given tags.
func (e *Entity) WithTags(tags ...string) *Entity {
	for i := range tags {
		e.tags[tags[i]] = struct{}{}
	}
	return e
}

// UID returns the unique identifier (Type() + ID()) for the Entity.
func (e *Entity) UID() string {
	return e.uid
}

// Type returns the type of Entity.
func (e *Entity) Type() string {
	return e.ns
}

// ID returns the identifier for the Entity.
func (e *Entity) ID() string {
	return e.id
}

// Title returns the title for the Entity.
func (e *Entity) Title() string {
	return e.title
}

// Description returns the description for the Entity.
func (e *Entity) Description() string {
	return e.description
}

// Attributes returns the attributes for the Entity.
func (e *Entity) Attributes() *AttributeSet {
	return e.attrs
}

// Parents returns the unique parents for the Entity.
func (e *Entity) Parents() []string {
	return e.parents
}

// Tags returns the tags for the Entity.
func (e *Entity) Tags() []string {
	tags := make([]string, 0, len(e.tags))
	for k := range e.tags {
		tags = append(tags, k)
	}
	slices.Sort(tags)
	return tags
}

// HasTag returns true if the Entity is annotated with the given tag.
func (e *Entity) HasTag(tag string) bool {
	_, ok := e.tags[tag]
	return ok
}

// MarshalJSON implements the json.Marshaler interface.
func (e *Entity) MarshalJSON() ([]byte, error) {
	return json.Marshal(marshalEntity{
		Type:        e.ns,
		ID:          e.id,
		Title:       e.title,
		Description: e.description,
		Attributes:  e.attrs,
		Parents:     e.parents,
		Tags:        e.Tags(),
	})
}

// MarshalYAML implements the yaml.Marshaler interface.
func (e *Entity) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(marshalEntity{
		Type:        e.ns,
		ID:          e.id,
		Title:       e.title,
		Description: e.description,
		Attributes:  e.attrs,
		Parents:     e.parents,
		Tags:        e.Tags(),
	})
}

// EntityToAttribute can be used to convert an entity into an attribute.
func EntityToAttribute(e *Entity) *Attribute {
	return NewAttributeWithType(e.UID(), mapFromEntity(e), "xsd:object")
}

func mapFromEntity(e *Entity) map[string]any {
	m := map[string]any{"type": e.Type(), "id": e.ID()}

	if attr := MapFromAttributes(e.Attributes()); len(attr) > 0 {
		m["attributes"] = attr
	}
	if parents := e.Parents(); len(parents) > 0 {
		m["parents"] = parents
	}

	return m
}

// EntityFromOAS instantiates a new Entity from the given OAS model.
func EntityFromOAS(in *attributes.Entity) *Entity {
	a := &Entity{
		uid:         EntityUID(in.Type, in.Id),
		ns:          in.Type,
		id:          in.Id,
		title:       in.Metadata.Title,
		description: in.Metadata.Description,
		attrs:       AttributeSetFromOAS(in.Attributes),
		tags:        make(map[string]struct{}, len(in.Metadata.Tags)),
	}

	for i := range in.Metadata.Tags {
		a.tags[in.Metadata.Tags[i]] = struct{}{}
	}

	return a
}

// ToOAS converts the Attribute to an OAS model.
func (e *Entity) ToOAS() *attributes.Entity {
	return &attributes.Entity{
		Type:       e.ns,
		Id:         e.id,
		Attributes: e.attrs.ToOAS(),
		Metadata: attributes.Metadata{
			Title:       e.title,
			Description: e.description,
			Tags:        e.Tags(),
		},
	}
}

// ToBundle converts the Attribute to an OAS model for bundling.
func (e *Entity) ToBundle() *attributes.Entity {
	return &attributes.Entity{
		Type:       e.ns,
		Id:         e.id,
		Attributes: e.attrs.ToOAS(),
	}
}

// Equals returns true if this Entity matches exactly with the other entity.
func (e *Entity) Equals(other *Entity) bool {
	return e.uid == other.uid &&
		e.ns == other.ns &&
		e.id == other.id &&
		e.title == other.title &&
		e.description == other.description &&
		reflect.DeepEqual(e.Parents(), other.Parents()) &&
		reflect.DeepEqual(e.Tags(), other.Tags()) &&
		e.Attributes().Equals(other.Attributes())
}

type marshalEntity struct {
	Type        string        `json:"type,omitempty"        yaml:"type,omitempty"`
	ID          string        `json:"id,omitempty"          yaml:"id,omitempty"`
	Title       string        `json:"title,omitempty"       yaml:"title,omitempty"`
	Description string        `json:"description,omitempty" yaml:"description,omitempty"`
	Attributes  *AttributeSet `json:"attributes,omitempty"  yaml:"attributes,omitempty"`
	Parents     []string      `json:"parents,omitempty"     yaml:"parents,omitempty"`
	Tags        []string      `json:"tags,omitempty"        yaml:"tags,omitempty"`
}
