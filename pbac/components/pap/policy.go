package pap

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/goccy/go-json"
	"gopkg.in/yaml.v3"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/policies"
)

// Policy represents a policy and its metadata.
//
// A policy is designed to be read-only, so it is safe to use across concurrent go-routines.
type Policy interface {
	ID() string
	Source() string
	Target() string
	RvvaID() string
	URI() string
	Path() string
	Content() io.Reader
}

// NewPolicy instantiates a new policy and its metadata.
func NewPolicy(id, source, target, rvvaID, uri string, content io.Reader) (Policy, error) {
	d, err := testContent(content)
	if err != nil {
		return nil, err
	}

	return &policy{
		id:      id,
		source:  source,
		target:  target,
		rvvaID:  rvvaID,
		uri:     uri,
		content: d,
	}, nil
}

// NewPolicyFromStore instantiates a new policy being loaded from the local store.
//
// The function will also attempt to load the policy's corresponding metadata from the same location.
// This file can be encoded as YAML or as JSON.
func NewPolicyFromStore(path string, content io.Reader) (Policy, error) {
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
			_ = json.NewDecoder(f).Decode(&p)
		}
	}

	if p.Id == "" {
		p.Id = id
	}

	return &policy{
		id:      p.Id,
		source:  p.Source,
		target:  p.Target,
		rvvaID:  p.RvvaID,
		uri:     p.Url,
		path:    path,
		content: d,
	}, nil
}

func testContent(content io.Reader) ([]byte, error) {
	if content == nil {
		return nil, errors.New("content is nil")
	}
	return io.ReadAll(content)
}

// ID implements the Policy interface.
func (p *policy) ID() string {
	return p.id
}

// Source implements the Policy interface.
func (p *policy) Source() string {
	return p.source
}

// Target implements the Policy interface.
func (p *policy) Target() string {
	return p.target
}

// RvvaID implements the Policy interface.
func (p *policy) RvvaID() string {
	return p.rvvaID
}

// URI implements the Policy interface.
func (p *policy) URI() string {
	return p.uri
}

// Path implements the Policy interface.
func (p *policy) Path() string {
	return p.path
}

// Content implements the Policy interface.
func (p *policy) Content() io.Reader {
	return bytes.NewReader(p.content)
}

type policy struct {
	id      string
	source  string
	target  string
	rvvaID  string
	uri     string
	path    string
	content []byte
}
