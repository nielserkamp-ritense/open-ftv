package rdf

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/xsd"

// FromString converts an RDF literal into a Golang variable of the appropriate type.
//
// This function supports the RDF data-types, as well as most of the common XSD data-types.
func FromString(data, t string) (any, error) {
	switch t {
	case HTML:
		return data, nil
	case XMLLiteral:
		return mapFromXML(data)
	case JSON:
		return mapFromJSON(data)
	default:
		return xsd.FromString(data, t)
	}
}
