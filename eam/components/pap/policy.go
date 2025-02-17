package pap

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/policies"
)

// Policy represents a policy and its metadata.
//
// A policy is designed to be read-only, so it is safe to use across concurrent go-routines.
type Policy interface {
	ID() string
	Language() string
	RvvaID() string
	URI() string
	Path() string
	Content() io.Reader
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

// NewPolicy instantiates a new policy from the given OAS model.
func NewPolicy(p *policies.Policy, content io.Reader) (Policy, error) {
	d, err := testContent(content)
	if err != nil {
		return nil, err
	}

	return newPolicy(p, "", d), nil
}

// NewPolicyFromData instantiates a new policy from the given details.
func NewPolicyFromData(id, language, rvvaID, uri string, content io.Reader) (Policy, error) {
	d, err := testContent(content)
	if err != nil {
		return nil, err
	}

	return &policy{
		id:       id,
		language: language,
		rvvaID:   rvvaID,
		uri:      uri,
		content:  d,
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
	return newPolicy(&p, path, d), nil
}

func newPolicy(p *policies.Policy, path string, d []byte) Policy {
	return &policy{
		id:       p.Id,
		language: p.Language,
		rvvaID:   p.RvvaId,
		uri:      p.Url,
		path:     path,
		content:  d,
	}
}

func testContent(content io.Reader) ([]byte, error) {
	if content == nil {
		return nil, errors.New("content is nil")
	}

	// TODO: protect against infinite input stream (DOS attack).
	return io.ReadAll(content)
}

// ID implements the Policy interface.
func (p *policy) ID() string {
	return p.id
}

// Language implements the Policy interface.
func (p *policy) Language() string {
	return p.language
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

// MarshalJSON implements the json.Marshaller interface.
func (p *policy) MarshalJSON() ([]byte, error) {
	return json.Marshal(&policyJSON{
		ID:       p.id,
		Language: p.language,
		RvvaID:   p.rvvaID,
		URI:      p.uri,
		Path:     p.path,
		Content:  base64.StdEncoding.EncodeToString(p.content),
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *policy) UnmarshalJSON(data []byte) error {
	p2 := &policyJSON{}
	if err := json.Unmarshal(data, p2); err != nil {
		return err
	}

	p.id = p2.ID
	p.language = p2.Language
	p.rvvaID = p2.RvvaID
	p.uri = p2.URI
	p.path = p2.Path
	p.content, _ = base64.StdEncoding.DecodeString(p2.Content)

	return nil
}

type policy struct {
	id       string
	language string
	rvvaID   string
	uri      string
	path     string
	content  []byte
}

type policyJSON struct {
	ID       string `json:"id"`
	Language string `json:"language"`
	RvvaID   string `json:"rvvaID,omitempty"`
	URI      string `json:"uri,omitempty"`
	Path     string `json:"path,omitempty"`
	Content  string `json:"content"`
}
