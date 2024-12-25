package io

// DefaultMimeType is the default mime-type, in case nothing else fits.
const DefaultMimeType = "application/octet-stream"

// list of supported mime types.
const (
	MimeTypeCSV       = "text/csv"
	MimeTypeCedar     = "application/x-policy-cedar"
	MimeTypeCerbos    = "application/x-policy-cerbos"
	MimeTypeJSON      = "application/json"
	MimeTypeJSONLD    = "application/ld+json"
	MimeTypeNotation3 = "text/n3"
	MimeTypeODRL      = "application/x-policy-odrl"
	MimeTypeOPA       = "application/x-policy-opa"
	MimeTypeOpenFGA   = "application/x-policy-openfga"
	MimeTypePlain     = "text/plain"
	MimeTypeRDF       = "application/rdf+xml"
	MimeTypeTOML      = "application/toml"
	MimeTypeTabSep    = "text/tab-separated-values"
	MimeTypeTurtle    = "text/turtle"
	MimeTypeXACML     = "application/x-policy-xacml"
	MimeTypeXML       = "application/xml"
	MimeTypeYAML      = "application/yaml"
)

// IsSupported returns true if the given mime-type is supported for decoding.
func IsSupported(t string) bool {
	_, ok := supportedTypes[t]
	return ok
}

// ConvertExt attempts to determine the mime-type based on the given file extension.
func ConvertExt(ext string) string {
	return extMap[ext]
}

var supportedTypes = map[string]struct{}{
	MimeTypeCedar:     {},
	MimeTypeCerbos:    {},
	MimeTypeCSV:       {},
	MimeTypeJSON:      {},
	MimeTypeJSONLD:    {},
	MimeTypeNotation3: {},
	MimeTypeODRL:      {},
	MimeTypeOPA:       {},
	MimeTypeOpenFGA:   {},
	MimeTypeRDF:       {},
	MimeTypeTabSep:    {},
	MimeTypeTOML:      {},
	MimeTypeTurtle:    {},
	MimeTypePlain:     {},
	MimeTypeXACML:     {},
	MimeTypeXML:       {},
	MimeTypeYAML:      {},
}

var extMap = map[string]string{
	".cedar":   MimeTypeCedar,
	".cerbos":  MimeTypeCerbos,
	".csv":     MimeTypeCSV,
	".json":    MimeTypeJSON,
	".json-ld": MimeTypeJSONLD,
	".jsonld":  MimeTypeJSONLD,
	".n3":      MimeTypeNotation3,
	".odrl":    MimeTypeODRL,
	".opa":     MimeTypeOPA,
	".openfga": MimeTypeOpenFGA,
	".rdf":     MimeTypeRDF,
	".rego":    MimeTypeOPA,
	".tab":     MimeTypeTabSep,
	".text":    MimeTypePlain,
	".tml":     MimeTypeTOML,
	".toml":    MimeTypeTOML,
	".tsv":     MimeTypeTabSep,
	".ttl":     MimeTypeTurtle,
	".turtle":  MimeTypeTurtle,
	".txt":     MimeTypePlain,
	".xacml":   MimeTypeXACML,
	".xml":     MimeTypeXML,
	".yaml":    MimeTypeYAML,
	".yml":     MimeTypeYAML,
}
