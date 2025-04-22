package network

import (
	"bytes"

	"github.com/deiu/rdf2go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/turtle"
)

func (r *runner) decodeRDF(data []byte, mime string) error {
	triples, err := turtle.Load(bytes.NewReader(data), mime)
	if err != nil {
		return err
	}
	return r.decodeTriples(triples)
}

func (r *runner) decodeTriples(_ *rdf2go.Graph) error {

	// TODO: iterate triple store and convert known FTV types to attributes, entities and/or relations.

	return nil
}
