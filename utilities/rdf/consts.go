package rdf

// List of RDF prefixes.
const (
	PrefixDCTerms = "dcterms"
	PrefixFTV     = "ftv"
	PrefixOWL     = "owl"
	PrefixRDF     = "rdf"
	PrefixRDFS    = "rdfs"
	PrefixSKOS    = "skos"
	PrefixXSD     = "xsd"
)

// List opf RDF uri's.
const (
	UriDCTerms = "http://purl.org/dc/terms/"
	UriFTV     = "https://ftv.nl/rdf/ftv#"
	UriOWL     = "http://www.w3.org/2002/07/owl#"
	UriRDF     = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	UriRDFS    = "http://www.w3.org/2000/01/rdf-schema#"
	UriSKOS    = "http://www.w3.org/2004/02/skos/core#"
	UriXSD     = "http://www.w3.org/2001/XMLSchema#"
)

// PrefixUri can be used to convert a well-known prefix into the corresponding URI.
var PrefixUri = map[string]string{
	PrefixDCTerms: UriDCTerms,
	PrefixFTV:     UriFTV,
	PrefixOWL:     UriOWL,
	PrefixRDF:     UriRDF,
	PrefixRDFS:    UriRDFS,
	PrefixSKOS:    UriSKOS,
	PrefixXSD:     UriXSD,
}

// UriPrefix can be used to convert a well-known URI into the corresponding prefix.
var UriPrefix = map[string]string{
	UriDCTerms: PrefixDCTerms,
	UriFTV:     PrefixFTV,
	UriOWL:     PrefixOWL,
	UriRDF:     PrefixRDF,
	UriRDFS:    PrefixRDFS,
	UriSKOS:    PrefixSKOS,
	UriXSD:     PrefixXSD,
}

// List of standard RDF identifiers.
const (
	Bag        = UriRDF + "Bag"
	HTML       = UriRDF + "HTML"
	JSON       = UriRDF + "JSON"
	List       = UriRDF + "List"
	Property   = UriRDF + "Property"
	Seq        = UriRDF + "Seq"
	Type       = UriRDF + "type"
	Value      = UriRDF + "value"
	XMLLiteral = UriRDF + "XMLLiteral"
)

// List of standard RDFS identifiers.
const (
	Class       = UriRDFS + "Class"
	Comment     = UriRDFS + "comment"
	DataType    = UriRDFS + "DataType"
	Domain      = UriRDFS + "domain"
	Label       = UriRDFS + "label"
	Literal     = UriRDFS + "Literal"
	Range       = UriRDFS + "range"
	Resource    = UriRDFS + "Resource"
	SubClass    = UriRDFS + "subClassOf"
	SubProperty = UriRDFS + "subPropertyOf"
)

// List of standard XSD identifiers.
const (
	XSDAny              = UriXSD + "anyType"
	XSDAnyURI           = UriXSD + "anyURI"
	XSDBoolean          = UriXSD + "boolean"
	XSDByte             = UriXSD + "byte"
	XSDDate             = UriXSD + "date"
	XSDDateTime         = UriXSD + "dateTime"
	XSDDay              = UriXSD + "gDay"
	XSDDecimal          = UriXSD + "decimal"
	XSDDouble           = UriXSD + "double"
	XSDDuration         = UriXSD + "duration"
	XSDFloat            = UriXSD + "float"
	XSDInt              = UriXSD + "int"
	XSDInteger          = UriXSD + "integer"
	XSDLanguage         = UriXSD + "language"
	XSDLong             = UriXSD + "long"
	XSDMonth            = UriXSD + "gMonth"
	XSDMonthDay         = UriXSD + "gMonthDay"
	XSDNeg              = UriXSD + "negativeInteger"
	XSDNonNeg           = UriXSD + "nonNegativeInteger"
	XSDNonPos           = UriXSD + "nonPositiveInteger"
	XSDNormalizedString = UriXSD + "normalizedString"
	XSDPos              = UriXSD + "positiveInteger"
	XSDShort            = UriXSD + "short"
	XSDSimple           = UriXSD + "simpleType"
	XSDString           = UriXSD + "string"
	XSDTime             = UriXSD + "time"
	XSDToken            = UriXSD + "token"
	XSDUByte            = UriXSD + "unsignedByte"
	XSDUInt             = UriXSD + "unsignedInt"
	XSDULong            = UriXSD + "unsignedLong"
	XSDUShort           = UriXSD + "unsignedShort"
	XSDYear             = UriXSD + "gYear"
	XSDYearMonth        = UriXSD + "gYearMonth"
)

// List of standard DC terms identifiers.
const (
	DCTermsDefined  = UriDCTerms + "isDefinedBy"
	DCTermsTitle    = UriDCTerms + "title"
	DCTermsRequired = UriDCTerms + "isRequiredBy"
	DCTermsRequires = UriDCTerms + "requires"
)

// List of standard OWL identifiers.
const (
	OWLClass = UriOWL + "Class"
	OWLThing = UriOWL + "Thing"
)

// List of standard SKOS identifiers.
const (
	SKOSAltLabel  = UriSKOS + "altLabel"
	SKOSConcept   = UriSKOS + "Concept"
	SKOSExample   = UriSKOS + "example"
	SKOSPrefLabel = UriSKOS + "prefLabel"
)

// List of FTV identifiers.
const (
	FTVAction          = UriFTV + "action"
	FTVAttribute       = UriFTV + "Attribute"
	FTVAttributeKey    = UriFTV + "attributeKey"
	FTVAttributeValue  = UriFTV + "attributeValue"
	FTVContext         = UriFTV + "context"
	FTVEntity          = UriFTV + "Entity"
	FTVEntityAttribute = UriFTV + "entityAttribute"
	FTVEntityID        = UriFTV + "entityID"
	FTVEntityType      = UriFTV + "entityType"
	FTVObject          = UriFTV + "object"
	FTVPrincipal       = UriFTV + "principal"
	FTVRelation        = UriFTV + "Relation"
	FTVRelationType    = UriFTV + "relationType"
	FTVResource        = UriFTV + "resource"
	FTVSubject         = UriFTV + "subject"
)
