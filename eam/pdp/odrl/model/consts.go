package model

// This file defines the RDF vocabulary constants that are specific to the
// ODRL-AP-NL application profile and the surrounding vocabularies (DCT, DCAT,
// PROV) that are used by the profile but not yet present in
// utilities/rdf/consts.go. The generic ODRL 2.2 constants are reused from
// utilities/rdf.

// APNLNamespace is the (work-version) namespace IRI of the ODRL-AP-NL profile.
const APNLNamespace = "https://standaarden.overheid.nl/odrl-ap-nl/"

// Namespace IRIs of application profiles that extend ODRL-AP-NL and are
// recognised (accepted without a warning) as additional odrl:profile values.
const (
	GeoNLProfile = "https://standaarden.overheid.nl/odrl-geo-nl/" // ODRL-Geo-NL spatial profile.
	BRPProfile   = "https://standaarden.overheid.nl/odrl-brp-nl/" // ODRL-BRP-NL profile.
)

// BRP profile vocabulary (odrl-brp-profiel.ttl).
const (
	// BRPDefNamespace is the brp: (def) namespace: BRP profile predicates and
	// LO-BRP operators (brp:knv, brp:ga1, …) live here.
	BRPDefNamespace = "https://data.rijksoverheid.nl/brp/def#"
	// BRPRubriekNamespace is the brprub: namespace of BRP field-groups (rubrieken).
	BRPRubriekNamespace = "https://data.rijksoverheid.nl/brp/rubriek/"

	// BRPVerzochteRubriek is brp:verzochteRubriek, an rdfs:subPropertyOf
	// odrl:target that makes the requested BRP field-groups (rubrieken) explicit
	// and machine-readable, so the RvIG harvester can derive the requested fields
	// (finding C1). The IRI is the one used by the BRP profile ontology.
	BRPVerzochteRubriek = BRPDefNamespace + "verzochteRubriek"
)

// ODRL-AP-NL classes and properties.
const (
	APNLProfile        = APNLNamespace // the profile IRI itself (odrl:profile target).
	APNLPolicyArtifact = APNLNamespace + "PolicyArtifact"
	APNLRegoModule     = APNLNamespace + "RegoModule"
	APNLCedarPolicySet = APNLNamespace + "CedarPolicySet"
	APNLOpenFGAModel   = APNLNamespace + "OpenFGAModel"
	APNLPolicyBundle   = APNLNamespace + "PolicyBundle"

	APNLBundles    = APNLNamespace + "bundles"
	APNLEntrypoint = APNLNamespace + "entrypoint"
	APNLSha256     = APNLNamespace + "sha256"

	APNLInstantiates = APNLNamespace + "instantiates"
	APNLWaivable     = APNLNamespace + "waivable"

	APNLVerwerkingsverzoek     = APNLNamespace + "verwerkingsverzoek"     // odrl:LeftOperand.
	APNLConformsToPolicy       = APNLNamespace + "conformsToPolicy"       // odrl:Operator.
	APNLDoelbindingsverwijzing = APNLNamespace + "doelbindingsverwijzing" // annotation.
)

// DCTerms vocabulary.
const (
	DCTNamespace   = "http://purl.org/dc/terms/"
	DCTTitle       = DCTNamespace + "title"
	DCTDescription = DCTNamespace + "description"
	DCTPublisher   = DCTNamespace + "publisher"
	DCTIssued      = DCTNamespace + "issued"
	DCTFormat      = DCTNamespace + "format"
	DCTConformsTo  = DCTNamespace + "conformsTo"
)

// DCAT vocabulary.
const (
	DCATNamespace    = "http://www.w3.org/ns/dcat#"
	DCATDataset      = DCATNamespace + "Dataset"
	DCATDistribution = DCATNamespace + "Distribution"
	DCATDownloadURL  = DCATNamespace + "downloadURL"
	DCATAccessURL    = DCATNamespace + "accessURL"
	DCATHasPolicy    = DCATNamespace + "hasPolicy"
	DCATDistProp     = DCATNamespace + "distribution"
)

// PROV vocabulary.
const (
	PROVNamespace      = "http://www.w3.org/ns/prov#"
	PROVWasDerivedFrom = PROVNamespace + "wasDerivedFrom"
	PROVWasRevisionOf  = PROVNamespace + "wasRevisionOf"
)

// RDFS vocabulary (subset used by named constraints/duties).
const (
	RDFSNamespace = "http://www.w3.org/2000/01/rdf-schema#"
	RDFSLabel     = RDFSNamespace + "label"
	RDFSComment   = RDFSNamespace + "comment"
)

// ODRL properties missing from utilities/rdf/consts.go (which only defines the
// capitalised class names for these).
const (
	odrlNS            = "http://www.w3.org/ns/odrl/2/"
	ODRLTargetProp    = odrlNS + "target"
	ODRLInformedParty = odrlNS + "informedParty"
)

// XSD datatypes used by literals.
const (
	XSDNamespace = "http://www.w3.org/2001/XMLSchema#"
	XSDBoolean   = XSDNamespace + "boolean"
	XSDDate      = XSDNamespace + "date"
	XSDString    = XSDNamespace + "string"
)

// ArtifactType maps a policy language to the ODRL-AP-NL PolicyArtifact subclass
// IRI and its dct:format media type. It is used by the export component.
type ArtifactType struct {
	Class     string // apnl class IRI.
	MediaType string // dct:format value.
}

// LanguageArtifact returns the ODRL-AP-NL artifact type for a policy language
// key (as used by the PAP, e.g. "rego"/"opa", "cedar", "openfga"). The second
// return value reports whether a mapping exists.
func LanguageArtifact(language string) (ArtifactType, bool) {
	switch language {
	case "rego", "opa", "opa-rego":
		return ArtifactType{Class: APNLRegoModule, MediaType: "application/vnd.rego"}, true
	case "cedar":
		return ArtifactType{Class: APNLCedarPolicySet, MediaType: "application/vnd.cedar"}, true
	case "openfga", "open-fga":
		return ArtifactType{Class: APNLOpenFGAModel, MediaType: "application/vnd.openfga"}, true
	default:
		return ArtifactType{}, false
	}
}

// ArtifactLanguage is the inverse of LanguageArtifact: it maps a dct:format
// media type or apnl artifact class IRI back to the PAP policy-language key.
func ArtifactLanguage(mediaTypeOrClass string) (string, bool) {
	switch mediaTypeOrClass {
	case "application/vnd.rego", APNLRegoModule:
		return "rego", true
	case "application/vnd.cedar", APNLCedarPolicySet:
		return "cedar", true
	case "application/vnd.openfga", APNLOpenFGAModel:
		return "openfga", true
	default:
		return "", false
	}
}
