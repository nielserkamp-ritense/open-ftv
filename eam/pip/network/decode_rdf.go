package network

import (
	"bytes"

	"github.com/deiu/rdf2go"
)

func (r *runner) decodeRDF(data []byte, mime string) error {
	triples := rdf2go.NewGraph("")
	if err := triples.Parse(bytes.NewReader(data), mime); err != nil {
		return err
	}
	return r.decodeTriples(triples)
}

func (r *runner) decodeTriples(_ *rdf2go.Graph) error {

	// TODO: iterate triple store and convert known FTV types to attributes, entities and/or relations.

	return nil
}
