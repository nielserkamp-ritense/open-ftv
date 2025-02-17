// Package rdf contains functionality for working with RDF data.
package rdf

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/xsd"

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
)

// List of well-known RDF uri's.
const (
	URICC      = "https://creativecommons.org/ns#"
	URIDCTerms = "http://purl.org/dc/terms/"
	URIFOAF    = "http://xmlns.com/foaf/0.1/"
	URIFTV     = "https://ftv.nl/rdf/ftv#"
	URIODRL    = "http://www.w3.org/ns/odrl/2/"
	URIOWL     = "http://www.w3.org/2002/07/owl#"
	URIRDF     = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	URIRDFS    = "http://www.w3.org/2000/01/rdf-schema#"
	URISKOS    = "http://www.w3.org/2004/02/skos/core#"
	URISchema  = "http://schema.org/"
	URIVCard   = "http://www.w3.org/2006/vcard/ns#"
)

// PrefixURI can be used to convert a well-known prefix into the corresponding URI.
var PrefixURI = map[string]string{
	PrefixCC:      URICC,
	PrefixDCT:     URIDCTerms,
	PrefixDCTerms: URIDCTerms,
	PrefixFOAF:    URIFOAF,
	PrefixFTV:     URIFTV,
	PrefixODRL:    URIODRL,
	PrefixOWL:     URIOWL,
	PrefixRDF:     URIRDF,
	PrefixRDFS:    URIRDFS,
	PrefixSKOS:    URISKOS,
	PrefixSchema:  URISchema,
	PrefixVCard:   URIVCard,
	xsd.Prefix:    xsd.URI,
}

// URIPrefix can be used to convert a well-known URI into the corresponding prefix.
var URIPrefix = map[string]string{
	URICC:      PrefixCC,
	URIDCTerms: PrefixDCTerms,
	URIFOAF:    PrefixFOAF,
	URIFTV:     PrefixFTV,
	URIODRL:    PrefixODRL,
	URIOWL:     PrefixOWL,
	URIRDF:     PrefixRDF,
	URIRDFS:    PrefixRDFS,
	URISKOS:    PrefixSKOS,
	URISchema:  PrefixSchema,
	URIVCard:   PrefixVCard,
	xsd.URI:    xsd.Prefix,
}

// List of standard RDF identifiers.
const (
	Bag        = URIRDF + "Bag"
	HTML       = URIRDF + "HTML"
	JSON       = URIRDF + "JSON"
	List       = URIRDF + "List"
	Property   = URIRDF + "Property"
	Seq        = URIRDF + "Seq"
	Type       = URIRDF + "type"
	Value      = URIRDF + "value"
	XMLLiteral = URIRDF + "XMLLiteral"
)

// List of standard RDFS identifiers.
const (
	Class       = URIRDFS + "Class"
	Comment     = URIRDFS + "comment"
	DataType    = URIRDFS + "DataType"
	Domain      = URIRDFS + "domain"
	Label       = URIRDFS + "label"
	Literal     = URIRDFS + "Literal"
	Range       = URIRDFS + "range"
	Resource    = URIRDFS + "Resource"
	SubClass    = URIRDFS + "subClassOf"
	SubProperty = URIRDFS + "subPropertyOf"
)

// List of standard DC terms identifiers.
const (
	DCTermsDefined  = URIDCTerms + "isDefinedBy"
	DCTermsTitle    = URIDCTerms + "title"
	DCTermsRequired = URIDCTerms + "isRequiredBy"
	DCTermsRequires = URIDCTerms + "requires"
)

// List of standard OWL identifiers.
const (
	OWLClass = URIOWL + "Class"
	OWLThing = URIOWL + "Thing"
)

// List of standard SKOS identifiers.
const (
	SKOSAltLabel  = URISKOS + "altLabel"
	SKOSConcept   = URISKOS + "Concept"
	SKOSExample   = URISKOS + "example"
	SKOSPrefLabel = URISKOS + "prefLabel"
)

// List of FTV identifiers.
const (
	FTVAction          = URIFTV + "action"
	FTVAttribute       = URIFTV + "Attribute"
	FTVAttributeKey    = URIFTV + "attributeKey"
	FTVAttributeValue  = URIFTV + "attributeValue"
	FTVContext         = URIFTV + "context"
	FTVEntity          = URIFTV + "Entity"
	FTVEntityAttribute = URIFTV + "entityAttribute"
	FTVEntityID        = URIFTV + "entityID"
	FTVEntityType      = URIFTV + "entityType"
	FTVObject          = URIFTV + "object"
	FTVPrincipal       = URIFTV + "principal"
	FTVRelation        = URIFTV + "Relation"
	FTVRelationType    = URIFTV + "relationType"
	FTVResource        = URIFTV + "resource"
	FTVSubject         = URIFTV + "subject"
)

// List of ODRL identifiers.
const (
	ODRLPolicy                = URIODRL + "Policy"
	ODRLIdentifier            = URIODRL + "uid"
	ODRLProfile               = URIODRL + "profile"
	ODRLInheritFrom           = URIODRL + "inheritFrom"
	ODRLAgreement             = URIODRL + "Agreement"
	ODRLOffer                 = URIODRL + "Offer"
	ODRLSet                   = URIODRL + "Set"
	ODRLRule                  = URIODRL + "Rule"
	ODRLRelation              = URIODRL + "relation"
	ODRLFunction              = URIODRL + "function"
	ODRLFailure               = URIODRL + "failure"
	ODRLAsset                 = URIODRL + "Asset"
	ODRLAssetCollection       = URIODRL + "AssetCollection"
	ODRLTarget                = URIODRL + "Target"
	ODRLTargetPolicy          = URIODRL + "TargetPolicy"
	ODRLParty                 = URIODRL + "Party"
	ODRLPartyCollection       = URIODRL + "PartyCollection"
	ODRLAssignee              = URIODRL + "assignee"
	ODRLAssigner              = URIODRL + "assigner"
	ODRLAssigneeOf            = URIODRL + "assigneeOf"
	ODRLAssignerOf            = URIODRL + "assignerOf"
	ODRLPartOf                = URIODRL + "partOf"
	ODRLSource                = URIODRL + "source"
	ODRLPermission            = URIODRL + "Permission"
	ODRLHasPermission         = URIODRL + "permission"
	ODRLProhibition           = URIODRL + "Prohibition"
	ODRLHasProhibition        = URIODRL + "prohibition"
	ODRLAction                = URIODRL + "Action"
	ODRLHasAction             = URIODRL + "action"
	ODRLIncludedIn            = URIODRL + "includedIn"
	ODRLImplies               = URIODRL + "implies"
	ODRLUse                   = URIODRL + "use"
	ODRLTransferOwnership     = URIODRL + "transfer"
	ODRLDuty                  = URIODRL + "Duty"
	ODRLHasDuty               = URIODRL + "duty"
	ODRLObligation            = URIODRL + "obligation"
	ODRLConsequence           = URIODRL + "consequence"
	ODRLRemedy                = URIODRL + "remedy"
	ODRLConstraint            = URIODRL + "Constraint"
	ODRLHasConstraint         = URIODRL + "constraint"
	ODRLRefinement            = URIODRL + "refinement"
	ODRLOperator              = URIODRL + "operator"
	ODRLRightOperand          = URIODRL + "RightOperand"
	ODRLHasRightOperand       = URIODRL + "rightOperand"
	ODRLHasRightOperandRef    = URIODRL + "rightOperandReference"
	ODRLLeftOperand           = URIODRL + "LeftOperand"
	ODRLHasLeftOperand        = URIODRL + "leftOperand"
	ODRLUnit                  = URIODRL + "unit"
	ODRLDatatype              = URIODRL + "datatype"
	ODRLStatus                = URIODRL + "status"
	ODRLLogicalConstraint     = URIODRL + "LogicalConstraint"
	ODRLOperand               = URIODRL + "operand"
	ODRLEqualTo               = URIODRL + "eq"
	ODRLGreaterThan           = URIODRL + "gt"
	ODRLGreaterThanOrEqual    = URIODRL + "gteq"
	ODRLLessThan              = URIODRL + "lt"
	ODRLLessThanOrEqual       = URIODRL + "lteq"
	ODRLNotEqualTo            = URIODRL + "neq"
	ODRLIsA                   = URIODRL + "isA"
	ODRLHasPart               = URIODRL + "hasPart"
	ODRLIsPartOf              = URIODRL + "isPartOf"
	ODRLIsAllOf               = URIODRL + "isAllOf"
	ODRLIsAnyOf               = URIODRL + "isAnyOf"
	ODRLIsNoneOf              = URIODRL + "isNoneOf"
	ODRLOr                    = URIODRL + "or"
	ODRLOnlyOne               = URIODRL + "xone"
	ODRLAnd                   = URIODRL + "and"
	ODRLAndSequence           = URIODRL + "andSequence"
	ODRLConflictStrategyPref  = URIODRL + "ConflictTerm"
	ODRLHandlePolicyConflicts = URIODRL + "conflict"
	ODRLPreferPermissions     = URIODRL + "perm"
	ODRLPreferProhibitions    = URIODRL + "prohibit"
	ODRLVoidPolicy            = URIODRL + "invalid"
)
