// Package bundles contains the functionality to handle policy&data bundles.
package bundles

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/dsnet/compress/bzip2"
	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// NewBundle instantiates a new bundle for policies and other data elements for pushing to one or more PDPs.
//
// A bundle is created with a specific version, policy language and one or more selection tags.
//
// The language and tags are used to select the appropriate policies for the bundle.
// For attributes, entities and/or relations only the tags are used for selection,
// as these elements are policy language agnostic.
//
// Note that only a single tag needs to match to select an element for the bundle.
//
// Once a bundle is instantiated, iterate over all policies, attributes, entities and relations,
// and pass them one by one to the Bundle.AddPolicy(), Bundle.AddAttribute(), etc. functions of the Bundle.
// The Bundle will automatically select the appropriate elements for inclusion.
//
// Once all elements have been processed, call the Bundle.Compress() function
// to write a JSON encoded, compressed binary bundle into the supplied io.Writer.
//
// When sending a binary blob over http, make sure to pass the appropriate Content-Encoding header,
// so the receiving handler will know the type of compression used.
//
// The version and language are included in an encoded bundle. The selection tags are not.
func NewBundle(version int, language string, tags ...string) *Bundle {
	b := newBundle()

	b.Version = version
	b.Language = language
	b.l = models.LanguageFromString(language)
	b.tags = append(b.tags, tags...)

	return b
}

// BundleFromAPI is used by a receiving PDP to decode a compressed binary bundle.
//
// The Content-Encoding header is used to determine the type of compression.
// If it is not supplied, or has no recognized value(s), gzip is assumed.
func BundleFromAPI(body io.Reader, headers map[string][]string) (*Bundle, error) {
	var ce CompressionType

	list := headers["Content-Encoding"]
	for i := range list {
		switch strings.ToLower(list[i]) {
		case "gzip", "gz":
			ce = CompressGZ
		case "bz2", "bzip2":
			ce = CompressBZ2
		}

		if ce > 0 {
			break
		}
	}

	switch ce {
	case CompressBZ2:
		return bundleFromBZip2(body)
	default:
		return bundleFromGZip(body)
	}
}

func bundleFromBZip2(body io.Reader) (b *Bundle, err error) {
	var r *bzip2.Reader
	if r, err = bzip2.NewReader(body, &bzip2.ReaderConfig{}); err == nil {
		b, err = bundleFromReader(r)
	}
	return
}

func bundleFromGZip(body io.Reader) (b *Bundle, err error) {
	var r *gzip.Reader
	if r, err = gzip.NewReader(body); err == nil {
		b, err = bundleFromReader(r)
	}
	return
}

func bundleFromReader(r io.ReadCloser) (b *Bundle, err error) {
	defer func() { err = errors.Join(err, r.Close()) }()

	b2 := newBundle()
	if err = json.NewDecoder(r).Decode(b2); err == nil {
		b = b2
	}
	return
}

func newBundle() *Bundle {
	return &Bundle{
		Policies:   make(map[string]*policies.Policy),
		Attributes: make(map[string]*attributes.Attribute),
		Entities:   make(map[string]*attributes.Entity),
		Relations:  make(map[string]*attributes.Relation),
		tags:       make([]string, 0),
	}
}

// Bundle represents a bundle of policies, attributes, entities and/or relations.
//
// It has a specific version number and a policy language to indicate the language of the included policies.
type Bundle struct {
	Version    int                              `json:"version"              yaml:"version"`
	Language   string                           `json:"language"             yaml:"language"`
	Policies   map[string]*policies.Policy      `json:"policies,omitempty"   yaml:"policies,omitempty"`
	Attributes map[string]*attributes.Attribute `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	Entities   map[string]*attributes.Entity    `json:"entities,omitempty"   yaml:"entities,omitempty"`
	Relations  map[string]*attributes.Relation  `json:"relations,omitempty"  yaml:"relations,omitempty"`
	// hidden fields.
	tags []string
	l    models.Language
}

// AddPolicy adds the given policy to the bundle if the language matches, and it contains one of the selection tags.
func (b *Bundle) AddPolicy(p pap.Policy) bool {
	if models.LanguageFromString(p.Language()) != b.l || !b.tagMatched(p.HasTag) {
		return false
	}

	buf, _ := io.ReadAll(p.Content())

	b.Policies[p.ID()] = &policies.Policy{Id: p.ID(), Data: string(buf)}
	return true
}

// AddAttribute adds the given attribute to the bundle if it contains one of the selection tags.
func (b *Bundle) AddAttribute(a *models.Attribute) bool {
	if !b.tagMatched(a.HasTag) {
		return false
	}

	b.Attributes[a.Key()] = &attributes.Attribute{Key: a.Key(), Type: a.Type(), Value: a.Value()}
	return true
}

// AddEntity adds the given entity to the bundle if it contains one of the selection tags.
func (b *Bundle) AddEntity(e *models.Entity) bool {
	if !b.tagMatched(e.HasTag) {
		return false
	}

	e2 := &attributes.Entity{Type: e.Type(), Id: e.ID(), Attributes: make([]attributes.Attribute, 0)}

	if list := e.Attributes(); list != nil {
		list.IterateAttributes(func(a *models.Attribute) {
			e2.Attributes = append(e2.Attributes, attributes.Attribute{Key: a.Key(), Type: a.Type(), Value: a.Value()})
		})
	}

	b.Entities[e.UID()] = e2
	return true
}

// AddRelation adds the given relation to the bundle if it contains one of the selection tags.
func (b *Bundle) AddRelation(r *models.Relation) bool {
	if !b.tagMatched(r.HasTag) {
		return false
	}

	r2 := &attributes.Relation{
		SubjectType: r.Subject().Type(),
		SubjectId:   r.Subject().ID(),
		Relation:    r.Predicate().ID(),
		ObjectType:  r.Object().Type(),
		ObjectId:    r.Object().ID(),
	}

	// TODO: attributes

	b.Relations[r.UID()] = r2
	return true
}

func (b *Bundle) tagMatched(f func(tag string) bool) bool {
	for i := range b.tags {
		if f(b.tags[i]) {
			return true
		}
	}
	return false
}

// Compress returns the bundle in JSON encoded compressed form.
func (b *Bundle) Compress(compress CompressionType, out io.Writer) error {
	switch compress {
	case CompressGZ:
		return b.compressGZ(out)
	case CompressBZ2:
		return b.compressBZ2(out)
	default:
		return fmt.Errorf("unsupported compression type: %v", compress)
	}
}

func (b *Bundle) compressGZ(out io.Writer) (err error) {
	var writer *gzip.Writer
	if writer, err = gzip.NewWriterLevel(out, gzip.BestCompression); err == nil {
		err = b.compress(writer)
	}
	return
}

func (b *Bundle) compressBZ2(out io.Writer) (err error) {
	var writer *bzip2.Writer
	if writer, err = bzip2.NewWriter(out, &bzip2.WriterConfig{Level: bzip2.BestCompression}); err == nil {
		err = b.compress(writer)
	}
	return
}

func (b *Bundle) compress(out io.WriteCloser) (err error) {
	defer func() { err = errors.Join(err, out.Close()) }()
	return json.NewEncoder(out).Encode(b)
}
