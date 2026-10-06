// Package turtle contains functionality for working with RDF files in Turtle format.
package turtle

import (
	"io"

	"github.com/deiu/rdf2go"
)

// Load loads a turtle graph from the given input stream.
//
// Use mime to indicate the mimetype for the input stream.
// Only "text/turtle" & "application/ld+json" are currently supported.
func Load(r io.Reader, mime string) (*rdf2go.Graph, error) {
	g := rdf2go.NewGraph("")
	if err := g.Parse(r, mime); err != nil {
		return nil, err
	}
	return g, nil
}
