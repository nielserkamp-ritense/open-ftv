package rdf

import (
	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/xsd"
)

// ConvertLiteral converts an RDF literal into a Golang variable of the appropriate type.
//
// This function supports the RDF data-types, as well as most of the common XSD data-types.
func ConvertLiteral(data, t string) (any, error) {
	switch t {
	case HTML:
		return data, nil
	case XMLLiteral:
		return convertXML(data)
	case JSON:
		return convertJSON(data)
	default:
		return xsd.Convert(data, t)
	}
}

// IsClass returns true if the given identifier can be identified as a class.
func IsClass(s string) bool {
	switch s {
	case Property, Class, Resource, DataType, OWLClass, OWLThing, SKOSConcept,
		FTVAttribute, FTVEntityAttribute, FTVEntity, FTVRelation,
		FTVPrincipal, FTVResource, FTVContext, FTVSubject, FTVObject:
		return true
	default:
		return false
	}
}

// IsLabel returns true if the given identifier can be identified as a label.
func IsLabel(s string) bool {
	switch s {
	case Label, SKOSPrefLabel, SKOSAltLabel:
		return true
	default:
		return false
	}
}

// IsTitle returns true if the given identifier can be identified as a title.
func IsTitle(s string) bool {
	switch s {
	case DCTermsTitle:
		return true
	default:
		return false
	}
}

// IsLiteral returns true if the given identifier can be identified as a literal.
func IsLiteral(s string) bool {
	switch s {
	case FTVAction, FTVAttributeKey, FTVEntityID, FTVEntityType, FTVRelationType,
		JSON, HTML, XMLLiteral,
		xsd.URIAny, xsd.URIAnyURI, xsd.URIBoolean, xsd.URIByte, xsd.URIDate, xsd.URIDateTime, xsd.URIDay, xsd.URIDecimal, xsd.URIDouble,
		xsd.URIDuration, xsd.URIFloat, xsd.URIInt, xsd.URIInteger, xsd.URILanguage, xsd.URILong, xsd.URIMonth,
		xsd.URIMonthDay, xsd.URINeg, xsd.URINonNeg, xsd.URINonPos, xsd.URINormalized, xsd.URIPos,
		xsd.URIShort, xsd.URISimple, xsd.URIString, xsd.URITime, xsd.URIToken,
		xsd.URIUByte, xsd.URIUInt, xsd.URIULong, xsd.URIUShort, xsd.URIYear, xsd.URIYearMonth:
		return true
	default:
		return false
	}
}

// IsList returns true if the given identifier can be identified as a list of items.
func IsList(s string) bool {
	switch s {
	case Bag, Seq, List:
		return true
	default:
		return false
	}
}

// IsAttribute returns true if the given identifier can be identified as an attribute.
func IsAttribute(s string) bool {
	switch s {
	case FTVAttribute, FTVEntityAttribute:
		return true
	default:
		return false
	}
}

// IsEntity returns true if the given identifier can be identified as an entity.
func IsEntity(s string) bool {
	switch s {
	case FTVEntity, FTVPrincipal, FTVResource, FTVContext, FTVSubject, FTVObject:
		return true
	default:
		return false
	}
}

// IsRelation returns true if the given identifier can be identified as a relation.
func IsRelation(s string) bool {
	switch s {
	case FTVRelation:
		return true
	default:
		return false
	}
}

// IsValue returns true if the given identifier can be identified as a value.
func IsValue(s string) bool {
	if IsLiteral(s) {
		return true
	}

	switch s {
	case Value, Property,
		FTVAttributeKey, FTVAttributeValue, FTVEntityType, FTVEntityID, FTVAction, FTVRelationType:
		return true
	default:
		return false
	}
}

func convertXML(data string) (any, error) {

	// TODO: extract XML data.

	return data, nil
}

func convertJSON(data string) (any, error) {
	var out any
	err := json.Unmarshal([]byte(data), &out)
	return out, err
}
