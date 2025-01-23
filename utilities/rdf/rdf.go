package rdf

import (
	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/xsd"
)

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
		JSON, HTML, XMLLiteral:
		return true
	default:
		return xsd.IsLiteral(s)
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

func mapFromXML(data string) (any, error) {
	// TODO: extract XML data.
	return data, nil
}

func mapFromJSON(data string) (any, error) {
	var out any
	err := json.Unmarshal([]byte(data), &out)
	return out, err
}
