package model

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/goccy/go-json"
)

// Envelope is the storage format for imported ODRL policies in the PAP store
// (policy language "odrl"): it keeps both the raw source document and the
// parsed model, so the original bytes can always be re-served.
type Envelope struct {
	UID      string    `json:"uid"`                // odrl:uid of the (first) policy.
	MimeType string    `json:"mimeType"`           // mime type of Source.
	Source   string    `json:"source"`             // base64 of the raw document.
	Document *Document `json:"document,omitempty"` // the parsed model.
}

// NewEnvelope wraps a raw ODRL document and its parsed model for storage.
func NewEnvelope(uid, mimeType string, source []byte, doc *Document) *Envelope {
	return &Envelope{
		UID:      uid,
		MimeType: mimeType,
		Source:   base64.StdEncoding.EncodeToString(source),
		Document: doc,
	}
}

// Bytes renders the envelope as JSON.
func (e *Envelope) Bytes() ([]byte, error) {
	return json.Marshal(e)
}

// RawSource returns the decoded raw source document.
func (e *Envelope) RawSource() ([]byte, error) {
	return base64.StdEncoding.DecodeString(e.Source)
}

// Mime types accepted for a raw (non-envelope) policy document.
const (
	mimeTurtle = "text/turtle"
	mimeJSONLD = "application/ld+json"
)

// ParseEnvelope reads a stored policy back from the PAP store. It accepts three
// on-disk shapes and always returns a fully-populated Envelope:
//
//   - a JSON envelope (as written by the importer);
//   - a raw Turtle/N3 document (a bare bronhouder beleid.ttl, as import.sh
//     copies it), detected because it does not start with '{';
//   - a raw JSON-LD document, detected because it starts with '{' but has no
//     envelope fields.
//
// Raw documents are parsed and wrapped in an envelope on the fly, so operators
// can drop a bare policy file into the store without hand-crafting JSON. When
// the envelope does not carry a parsed document, the raw source is re-parsed.
func ParseEnvelope(data []byte) (*Envelope, error) {
	trimmed := bytes.TrimLeft(data, " \t\r\n")
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("odrl: empty policy document")
	}

	// A raw RDF document: Turtle/N3 starts with '@' (prefix), '#' (comment) or a
	// keyword (PREFIX/BASE), never with '{'.
	if trimmed[0] != '{' {
		return wrapRaw(data, mimeTurtle)
	}

	e := &Envelope{}
	if err := json.Unmarshal(data, e); err != nil {
		// A '{' that is not valid envelope JSON: try raw JSON-LD.
		return wrapRaw(data, mimeJSONLD)
	}
	if e.Source == "" && e.Document == nil {
		// Valid JSON but not an envelope (no source, no parsed model): raw JSON-LD.
		return wrapRaw(data, mimeJSONLD)
	}

	if e.Document == nil {
		src, err := e.RawSource()
		if err != nil {
			return nil, fmt.Errorf("odrl: invalid envelope source: %w", err)
		}
		doc, err2 := Parse(bytes.NewReader(src), e.MimeType)
		if err2 != nil {
			return nil, err2
		}
		e.Document = doc
	}

	return e, nil
}

// wrapRaw parses a raw ODRL document (Turtle or JSON-LD) and wraps it in an
// envelope so it re-serves identically and exposes a parsed model.
func wrapRaw(data []byte, mimeType string) (*Envelope, error) {
	doc, err := Parse(bytes.NewReader(data), mimeType)
	if err != nil {
		return nil, fmt.Errorf("odrl: cannot parse raw policy document: %w", err)
	}
	uid := ""
	if len(doc.Policies) > 0 {
		uid = doc.Policies[0].UID
	}
	return NewEnvelope(uid, mimeType, data, doc), nil
}
