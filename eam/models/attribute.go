// Package models defines generic constants, enumerations and models used for EAM.
package models

import (
	"reflect"
	"slices"
	"sync"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"golang.org/x/exp/maps"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

// Attribute contains the details of an attribute.
//
// An Attribute is safe for use in concurrent go-routines.
type Attribute struct {
	key         string
	title       string
	description string
	value       any
	original    any
	tp          string
	tags        map[string]struct{}
	mutex       sync.RWMutex
}

// NewAttribute instantiates a new Attribute without a specific type.
func NewAttribute(key string, value any) *Attribute {
	return NewOriginalAttribute(key, value, value, "")
}

// NewAttributeWithType instantiates a new Attribute with a specific type.
func NewAttributeWithType(key string, value any, tp string) *Attribute {
	return NewOriginalAttribute(key, value, value, tp)
}

// NewOriginalAttribute instantiates a new Attribute with a specific type and original value.
func NewOriginalAttribute(key string, value, original any, tp string) *Attribute {
	return &Attribute{key: key, value: valueFromOAS(value, tp), original: original, tp: tp, tags: make(map[string]struct{})}
}

// NewAttributeFromOAS instantiates a new Attribute from the given OAS model.
func NewAttributeFromOAS(in *attributes.Attribute) *Attribute {
	a := &Attribute{
		key:         in.Key,
		title:       in.Metadata.Title,
		description: in.Metadata.Description,
		value:       valueFromOAS(in.Value, in.Type),
		original:    in.Value,
		tp:          in.Type,
		tags:        make(map[string]struct{}, len(in.Metadata.Tags)),
	}

	for i := range in.Metadata.Tags {
		a.tags[in.Metadata.Tags[i]] = struct{}{}
	}

	return a
}

// WithTitle adds an optional title to the Attribute.
func (a *Attribute) WithTitle(title string) *Attribute {
	a.mutex.Lock()
	a.title = title
	a.mutex.Unlock()
	return a
}

// WithDescription adds an optional description to the Attribute.
func (a *Attribute) WithDescription(desc string) *Attribute {
	a.mutex.Lock()
	a.description = desc
	a.mutex.Unlock()
	return a
}

// WithTags annotates the Attribute with the given tags.
func (a *Attribute) WithTags(tags ...string) *Attribute {
	a.mutex.Lock()
	for i := range tags {
		a.tags[tags[i]] = struct{}{}
	}
	a.mutex.Unlock()
	return a
}

// Key returns the key of the Attribute.
func (a *Attribute) Key() string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.key
}

// Title returns the title of the Attribute.
func (a *Attribute) Title() string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.title
}

// Description returns the description of the Attribute.
func (a *Attribute) Description() string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.description
}

// Value returns the value of the Attribute.
func (a *Attribute) Value() any {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.value
}

// Original returns the original value of the Attribute.
func (a *Attribute) Original() any {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.original
}

// Type returns the type of Attribute.
func (a *Attribute) Type() string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.tp
}

// Tags returns the tags for the Attribute.
func (a *Attribute) Tags() []string {
	a.mutex.RLock()
	tags := maps.Keys(a.tags)
	a.mutex.RUnlock()

	slices.Sort(tags)
	return tags
}

// HasTag returns true if the Attribute is annotated with the given tag.
func (a *Attribute) HasTag(tag string) bool {
	a.mutex.RLock()
	_, ok := a.tags[tag]
	a.mutex.RUnlock()
	return ok
}

// MarshalJSON implements the json.Marshaler interface.
func (a *Attribute) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.newMarshallAttr())
}

// MarshalYAML implements the yaml.Marshaler interface.
func (a *Attribute) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(a.newMarshallAttr())
}

// ToOAS converts the Attribute to an OAS model.
func (a *Attribute) ToOAS() *attributes.Attribute {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	value, tp := valueToOAS(a.value, a.tp)

	return &attributes.Attribute{
		Key:   a.key,
		Value: value,
		Type:  tp,
		Metadata: attributes.Metadata{
			Title:       a.title,
			Description: a.description,
			Tags:        a.Tags(),
		},
	}
}

// ToBundle converts the Attribute to an OAS model for bundling.
func (a *Attribute) ToBundle() *attributes.Attribute {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	value, tp := valueToOAS(a.value, a.tp)

	return &attributes.Attribute{
		Key:   a.key,
		Value: value,
		Type:  tp,
	}
}

// Equals returns true if this Attribute equals the other Attribute.
func (a *Attribute) Equals(other *Attribute) bool {
	return a.key == other.key &&
		a.title == other.title &&
		a.description == other.description &&
		a.tp == other.tp &&
		reflect.DeepEqual(a.value, other.value) &&
		reflect.DeepEqual(a.original, other.original) &&
		reflect.DeepEqual(a.tags, other.tags)
}

func (a *Attribute) newMarshallAttr() *marshalAttr {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	m := &marshalAttr{
		Key:         a.key,
		Title:       a.title,
		Description: a.description,
		Value:       a.value,
		Type:        a.tp,
		Tags:        a.Tags(),
	}

	if a.original != a.value {
		m.Original = a.original
	}

	return m
}

type marshalAttr struct {
	Key         string   `json:"key"                   yaml:"key"`
	Title       string   `json:"title,omitempty"       yaml:"title,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Value       any      `json:"value"                 yaml:"value"`
	Original    any      `json:"original,omitempty"    yaml:"original,omitempty"`
	Type        string   `json:"type,omitempty"        yaml:"type,omitempty"`
	Tags        []string `json:"tags,omitempty"        yaml:"tags,omitempty"`
}
