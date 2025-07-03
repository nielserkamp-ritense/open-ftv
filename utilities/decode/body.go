// Package decode contains functionality for decoding data in various formats.
package decode

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// ParseBody parses a body blob into a map of key/value pairs.
//
// This given content-type is used to determine which decoder to use.
// If the given content-type is empty or not recognized,
// the content-type will be determined based on the data.
//
// Currently supported content-types are JSON and XML.
func ParseBody(body []byte, contentType string) (map[string]any, error) {
	ct := strings.ToLower(convert.RemoveHeaderParameters(contentType))

	parser, ok := parsers[ct]
	if !ok {
		ct = convert.ContentSniffer(body)
		parser, ok = parsers[ct]
	}

	if !ok {
		return nil, fmt.Errorf("no parser found for content-type %s", contentType)
	}
	return parser(body)
}

type bodyParser func(body []byte) (map[string]any, error)

var parsers = map[string]bodyParser{
	"text/xml":              parseXML,
	"application/xml":       parseXML,
	"application/soap+xml":  parseXML,
	"application/atom+xml":  parseXML,
	"application/rss+xml":   parseXML,
	"application/rdf+xml":   parseXML,
	"application/xhtml+xml": parseXML,
	"application/xslt+xml":  parseXML,
	"text/json":             parseJSON,
	"application/json":      parseJSON,
	"application/ld+json":   parseJSON,
	"application/geo+json":  parseJSON,
}

func parseJSON(body []byte) (map[string]any, error) {
	m := make(map[string]any)
	if err := json.NewDecoder(bytes.NewBuffer(body)).Decode(&m); err != nil && err != io.EOF {
		return nil, err
	}
	return m, nil
}

func parseXML(body []byte) (map[string]any, error) {
	nodes := make([]xmlNode, 0)
	if err := xml.Unmarshal(body, &nodes); err != nil {
		return nil, err
	}

	nodes2 := make([]map[string]any, len(nodes))
	for i := range nodes {
		node := nodes[i]
		nodes2[i] = map[string]any{node.name(): node.toMap()}
	}

	return map[string]any{"nodes": nodes2}, nil
}
