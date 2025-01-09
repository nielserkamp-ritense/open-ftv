package rdf

// List of well-known RDF prefixes.
const (
	PrefixCC      = "cc"
	PrefixDCT     = "dct"
	PrefixDCTerms = "dcterms"
	PrefixFOAF    = "foaf"
	PrefixFTV     = "ftv"
	PrefixODRL    = "odrl"
	PrefixOWL     = "owl"
	PrefixRDF     = "rdf"
	PrefixRDFS    = "rdfs"
	PrefixSKOS    = "skos"
	PrefixSchema  = "schema"
	PrefixVCard   = "vcard"
	PrefixXSD     = "xsd"
)

// List of well-known RDF uri's.
const (
	UriCC      = "https://creativecommons.org/ns#"
	UriDCTerms = "http://purl.org/dc/terms/"
	UriFOAF    = "http://xmlns.com/foaf/0.1/"
	UriFTV     = "https://ftv.nl/rdf/ftv#"
	UriODRL    = "http://www.w3.org/ns/odrl/2/"
	UriOWL     = "http://www.w3.org/2002/07/owl#"
	UriRDF     = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	UriRDFS    = "http://www.w3.org/2000/01/rdf-schema#"
	UriSKOS    = "http://www.w3.org/2004/02/skos/core#"
	UriSchema  = "http://schema.org/"
	UriVCard   = "http://www.w3.org/2006/vcard/ns#"
	UriXSD     = "http://www.w3.org/2001/XMLSchema#"
)

// PrefixUri can be used to convert a well-known prefix into the corresponding URI.
var PrefixUri = map[string]string{
	PrefixCC:      UriCC,
	PrefixDCT:     UriDCTerms,
	PrefixDCTerms: UriDCTerms,
	PrefixFOAF:    UriFOAF,
	PrefixFTV:     UriFTV,
	PrefixODRL:    UriODRL,
	PrefixOWL:     UriOWL,
	PrefixRDF:     UriRDF,
	PrefixRDFS:    UriRDFS,
	PrefixSKOS:    UriSKOS,
	PrefixSchema:  UriSchema,
	PrefixVCard:   UriVCard,
	PrefixXSD:     UriXSD,
}

// UriPrefix can be used to convert a well-known URI into the corresponding prefix.
var UriPrefix = map[string]string{
	UriCC:      PrefixCC,
	UriDCTerms: PrefixDCTerms,
	UriFOAF:    PrefixFOAF,
	UriFTV:     PrefixFTV,
	UriODRL:    PrefixODRL,
	UriOWL:     PrefixOWL,
	UriRDF:     PrefixRDF,
	UriRDFS:    PrefixRDFS,
	UriSKOS:    PrefixSKOS,
	UriSchema:  PrefixSchema,
	UriVCard:   PrefixVCard,
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

// List of ODRL identifiers.
const (
	ODRLPolicy                = UriODRL + "Policy"
	ODRLIdentifier            = UriODRL + "uid"
	ODRLProfile               = UriODRL + "profile"
	ODRLInheritFrom           = UriODRL + "inheritFrom"
	ODRLAgreement             = UriODRL + "Agreement"
	ODRLOffer                 = UriODRL + "Offer"
	ODRLSet                   = UriODRL + "Set"
	ODRLRule                  = UriODRL + "Rule"
	ODRLRelation              = UriODRL + "relation"
	ODRLFunction              = UriODRL + "function"
	ODRLFailure               = UriODRL + "failure"
	ODRLAsset                 = UriODRL + "Asset"
	ODRLAssetCollection       = UriODRL + "AssetCollection"
	ODRLTarget                = UriODRL + "Target"
	ODRLTargetPolicy          = UriODRL + "TargetPolicy"
	ODRLParty                 = UriODRL + "Party"
	ODRLPartyCollection       = UriODRL + "PartyCollection"
	ODRLAssignee              = UriODRL + "assignee"
	ODRLAssigner              = UriODRL + "assigner"
	ODRLAssigneeOf            = UriODRL + "assigneeOf"
	ODRLAssignerOf            = UriODRL + "assignerOf"
	ODRLPartOf                = UriODRL + "partOf"
	ODRLSource                = UriODRL + "source"
	ODRLPermission            = UriODRL + "Permission"
	ODRLHasPermission         = UriODRL + "permission"
	ODRLProhibition           = UriODRL + "Prohibition"
	ODRLHasProhibition        = UriODRL + "prohibition"
	ODRLAction                = UriODRL + "Action"
	ODRLHasAction             = UriODRL + "action"
	ODRLIncludedIn            = UriODRL + "includedIn"
	ODRLImplies               = UriODRL + "implies"
	ODRLUse                   = UriODRL + "use"
	ODRLTransferOwnership     = UriODRL + "transfer"
	ODRLDuty                  = UriODRL + "Duty"
	ODRLHasDuty               = UriODRL + "duty"
	ODRLObligation            = UriODRL + "obligation"
	ODRLConsequence           = UriODRL + "consequence"
	ODRLRemedy                = UriODRL + "remedy"
	ODRLConstraint            = UriODRL + "Constraint"
	ODRLHasConstraint         = UriODRL + "constraint"
	ODRLRefinement            = UriODRL + "refinement"
	ODRLOperator              = UriODRL + "operator"
	ODRLRightOperand          = UriODRL + "RightOperand"
	ODRLHasRightOperand       = UriODRL + "rightOperand"
	ODRLHasRightOperandRef    = UriODRL + "rightOperandReference"
	ODRLLeftOperand           = UriODRL + "LeftOperand"
	ODRLHasLeftOperand        = UriODRL + "leftOperand"
	ODRLUnit                  = UriODRL + "unit"
	ODRLDatatype              = UriODRL + "datatype"
	ODRLStatus                = UriODRL + "status"
	ODRLLogicalConstraint     = UriODRL + "LogicalConstraint"
	ODRLOperand               = UriODRL + "operand"
	ODRLEqualTo               = UriODRL + "eq"
	ODRLGreaterThan           = UriODRL + "gt"
	ODRLGreaterThanOrEqual    = UriODRL + "gteq"
	ODRLLessThan              = UriODRL + "lt"
	ODRLLessThanOrEqual       = UriODRL + "lteq"
	ODRLNotEqualTo            = UriODRL + "neq"
	ODRLIsA                   = UriODRL + "isA"
	ODRLHasPart               = UriODRL + "hasPart"
	ODRLIsPartOf              = UriODRL + "isPartOf"
	ODRLIsAllOf               = UriODRL + "isAllOf"
	ODRLIsAnyOf               = UriODRL + "isAnyOf"
	ODRLIsNoneOf              = UriODRL + "isNoneOf"
	ODRLOr                    = UriODRL + "or"
	ODRLOnlyOne               = UriODRL + "xone"
	ODRLAnd                   = UriODRL + "and"
	ODRLAndSequence           = UriODRL + "andSequence"
	ODRLConflictStrategyPref  = UriODRL + "ConflictTerm"
	ODRLHandlePolicyConflicts = UriODRL + "conflict"
	ODRLPreferPermissions     = UriODRL + "perm"
	ODRLPreferProhibitions    = UriODRL + "prohibit"
	ODRLVoidPolicy            = UriODRL + "invalid"
)
