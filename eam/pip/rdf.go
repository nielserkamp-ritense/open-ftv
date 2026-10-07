package pip

import (
	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/xsd"
)

const (
	uriRDF = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	uriFTV = "https://ftv.nl/rdf/ftv#"
)

const (
	rdfHTML       = uriRDF + "HTML"
	rdfJSON       = uriRDF + "JSON"
	rdfType       = uriRDF + "type"
	rdfXMLLiteral = uriRDF + "XMLLiteral"
)

const (
	ftvAttribute       = uriFTV + "Attribute"
	ftvAttributeKey    = uriFTV + "attributeKey"
	ftvAttributeValue  = uriFTV + "attributeValue"
	ftvEntity          = uriFTV + "Entity"
	ftvEntityAttribute = uriFTV + "entityAttribute"
	ftvEntityID        = uriFTV + "entityID"
	ftvEntityType      = uriFTV + "entityType"
)

func rdfFromString(data, t string) (any, error) {
	switch t {
	case rdfHTML, rdfXMLLiteral:
		return data, nil
	case rdfJSON:
		var out any

		err := json.Unmarshal([]byte(data), &out)

		return out, err
	default:
		return xsd.FromString(data, t)
	}
}
