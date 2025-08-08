package models

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"golang.org/x/exp/maps"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// PolicyIterator is the function prototype to iterate through a set of policies.
type PolicyIterator func(p *Policy)

// Policy represents a policy and its metadata.
//
// A policy is safe to use across concurrent go-routines.
type Policy struct {
	language    string
	id          string
	title       string
	description string
	tags        map[string]struct{}
	rvvaID      string
	uri         string
	path        string
	content     []byte
	mutex       sync.RWMutex
}

// NewPolicyFromOAS instantiates a new policy from the given OAS model.
func NewPolicyFromOAS(p *policies.Policy, content io.Reader) (*Policy, error) {
	d, err := testContent(content)
	if err != nil {
		return nil, err
	}
	return newPolicy(p, "", d), nil
}

// NewPolicyFromStore instantiates a new policy being loaded from the local store.
//
// The function will also attempt to load the policy's corresponding metadata from the same location.
// The metadata file can be encoded as YAML or as JSON.
func NewPolicyFromStore(language, path string, content io.Reader) (*Policy, error) {
	d, err := testContent(content)
	if err != nil {
		return nil, err
	}

	id, ext := filepath.Base(path), filepath.Ext(path)
	meta := path[:len(path)-len(ext)] + ".meta"

	var p policies.Policy
	if f, err2 := os.Open(meta); err2 == nil {
		defer f.Close()
		if err2 = yaml.NewDecoder(f).Decode(&p); err2 != nil {
			if err2 = json.NewDecoder(f).Decode(&p); err2 != nil {
				return nil, err2
			}
		}
	}

	if p.Language == "" {
		p.Language = language
	}
	if p.Id == "" {
		p.Id = id
	}

	return newPolicy(&p, path, d), nil
}

// NewPolicyFromData instantiates a new policy from the given details.
func NewPolicyFromData(id, language, rvvaID, uri string, content io.Reader) (*Policy, error) {
	d, err := testContent(content)
	if err != nil {
		return nil, err
	}

	return &Policy{
		id:       id,
		language: strings.ToLower(language),
		rvvaID:   rvvaID,
		uri:      uri,
		content:  d,
		tags:     make(map[string]struct{}),
	}, nil
}

func newPolicy(p *policies.Policy, path string, d []byte) *Policy {
	tags := make(map[string]struct{})
	for i := range p.Metadata.Tags {
		tags[p.Metadata.Tags[i]] = struct{}{}
	}

	return &Policy{
		id:          p.Id,
		title:       p.Metadata.Title,
		description: p.Metadata.Description,
		tags:        tags,
		language:    strings.ToLower(p.Language),
		rvvaID:      p.Metadata.RvvaId,
		uri:         p.Metadata.Url,
		path:        path,
		content:     d,
	}
}

func testContent(content io.Reader) ([]byte, error) {
	if content == nil {
		return nil, errors.New("content is nil")
	}
	return io.ReadAll(content)
}

// WithTitle adds an optional title to the Policy.
func (p *Policy) WithTitle(title string) *Policy {
	p.mutex.Lock()
	p.title = title
	p.mutex.Unlock()
	return p
}

// WithDescription adds an optional description to the Policy.
func (p *Policy) WithDescription(desc string) *Policy {
	p.mutex.Lock()
	p.description = desc
	p.mutex.Unlock()
	return p
}

// WithTags annotates the Policy with the given tags.
func (p *Policy) WithTags(tags ...string) *Policy {
	p.mutex.Lock()
	for i := range tags {
		p.tags[tags[i]] = struct{}{}
	}
	p.mutex.Unlock()
	return p
}

// Key returns the unique key of the Policy.
func (p *Policy) Key() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return fmt.Sprintf("%s/%s", strings.ToLower(p.language), p.id)
}

// Language returns the language of the Policy.
//
// Note that Language also serves as a tag.
func (p *Policy) Language() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.language
}

// ID returns the unique ID of the Policy.
func (p *Policy) ID() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.id
}

// Title returns the title of the Policy.
func (p *Policy) Title() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.title
}

// Description returns the description of the Policy.
func (p *Policy) Description() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.description
}

// Tags returns the tags for the Policy.
func (p *Policy) Tags() []string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	tags := make([]string, 0, len(p.tags))
	for k := range p.tags {
		tags = append(tags, k)
	}
	slices.Sort(tags)
	return tags
}

// HasTag returns true if the Policy is annotated with the given tag.
//
// Note that the Language and RvvaID are also tags.
func (p *Policy) HasTag(tag string) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if strings.EqualFold(p.language, tag) || strings.EqualFold(p.rvvaID, tag) {
		return true
	}

	_, ok := p.tags[tag]
	return ok
}

// RvvaID returns the unique *Register van Verwerkings Activiteiten* identifier for the Policy.
//
// Note that RvvaID also serves as a tag.
func (p *Policy) RvvaID() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.rvvaID
}

// URI returns the URI of the Policy.
func (p *Policy) URI() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.uri
}

// Path returns the location on disk where the Policy was obtained from.
func (p *Policy) Path() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.path
}

// Content returns an io.Reader of the Policy content.
func (p *Policy) Content() io.Reader {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return bytes.NewReader(p.content)
}

// ToOAS returns the OAS model for this Policy.
func (p *Policy) ToOAS(withData bool) *policies.Policy {
	out := &policies.Policy{
		Id:       p.ID(),
		Language: p.Language(),
		Metadata: policies.Metadata{
			Title:       p.Title(),
			Description: p.Description(),
			RvvaId:      p.RvvaID(),
			Tags:        p.Tags(),
		},
	}

	if !withData || p.URI() != "" {
		out.Metadata.Url = p.URI()
		return out
	}

	out.Data = string(p.content)
	return out
}

// ToBundle returns the bundle model for this Policy.
func (p *Policy) ToBundle() *policies.Policy {
	return &policies.Policy{
		Id:       p.ID(),
		Language: p.Language(),
		Data:     string(p.content),
	}
}

// MarshalJSON implements the json.Marshaler interface.
func (p *Policy) MarshalJSON() ([]byte, error) {
	m := p.marshal()
	return json.Marshal(m)
}

// MarshalYAML implements the yaml.Marshaler interface.
func (p *Policy) MarshalYAML() ([]byte, error) {
	m := p.marshal()
	return yaml.Marshal(m)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *Policy) UnmarshalJSON(data []byte) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p2 := &marshalPolicy{}
	if err := json.Unmarshal(data, p2); err != nil {
		return err
	}

	p.language = strings.ToLower(p2.Language)
	p.id = p2.ID
	p.title = p2.Title
	p.description = p2.Description
	p.rvvaID = p2.RvvaID
	p.uri = p2.URI
	p.path = p2.Path
	p.content, _ = base64.StdEncoding.DecodeString(p2.Content)

	if p.tags == nil {
		p.tags = make(map[string]struct{}, len(p2.Tags))
	} else {
		maps.Clear(p.tags)
	}

	for i := range p2.Tags {
		p.tags[p2.Tags[i]] = struct{}{}
	}

	return nil
}

// SplitPolicyKey splits a policy-key into language and id.
func SplitPolicyKey(key string) (string, string) {
	parts := strings.Split(key, "/")
	switch len(parts) {
	case 2:
		return parts[0], parts[1]
	case 3:
		if strings.EqualFold(parts[0], "opa") || strings.EqualFold(parts[0], "cerbos") {
			return strings.Join(parts[0:2], "/"), parts[2]
		}
	}

	return "", key
}

func (p *Policy) marshal() *marshalPolicy {
	tags := make([]string, 0, len(p.tags))

	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for k := range p.tags {
		tags = append(tags, k)
	}
	slices.Sort(tags)

	return &marshalPolicy{
		Language:    p.language,
		ID:          p.id,
		Title:       p.title,
		Description: p.description,
		Tags:        tags,
		RvvaID:      p.rvvaID,
		URI:         p.uri,
		Path:        p.path,
		Content:     base64.StdEncoding.EncodeToString(p.content),
	}
}

func (p *Policy) Equals(other *Policy) bool {
	return p.language == other.language &&
		p.title == other.title &&
		p.description == other.description &&
		p.rvvaID == other.rvvaID &&
		p.uri == other.uri &&
		p.path == other.path &&
		bytes.Equal(p.content, other.content) &&
		reflect.DeepEqual(p.tags, other.tags)
}

type marshalPolicy struct {
	Language    string   `json:"language"              yaml:"language"`
	ID          string   `json:"id"                    yaml:"id"`
	Title       string   `json:"title,omitempty"       yaml:"title,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"        yaml:"tags,omitempty"`
	RvvaID      string   `json:"rvvaID,omitempty"      yaml:"rvvaID,omitempty"`
	URI         string   `json:"uri,omitempty"         yaml:"uri,omitempty"`
	Path        string   `json:"path,omitempty"        yaml:"path,omitempty"`
	Content     string   `json:"content"               yaml:"content"`
}
