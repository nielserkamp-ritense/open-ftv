package models

import "strings"

// Format represents the format for policy files.
type Format uint8

// List of supported policy file formats.
//
// Note that the policy language (Language) limits the choices for the file format of a policy.
const (
	XML Format = iota + 1
	RDFXML
	Turtle
	N3
	JSONLD
	JSON
	YAML
	FreeFormat

	fmtFirst = XML
	fmtLast  = FreeFormat
)

// String implements the Stringer interface.
func (f Format) String() string {
	switch f {
	case XML:
		return "XML"
	case RDFXML:
		return "RDF/XML"
	case Turtle:
		return "Turtle"
	case N3:
		return "N3"
	case JSONLD:
		return "JSON-LD"
	case JSON:
		return "JSON"
	case YAML:
		return "YAML"
	case FreeFormat:
		return "free-format"
	default:
		return "<unknown>"
	}
}

// FormatFromString returns the corresponding Format type from the given input.
func FormatFromString(in string) Format {
	return formats[strings.ToLower(in)]
}

var formats = make(map[string]Format, fmtLast+4)

func init() {
	for i := fmtFirst; i <= fmtLast; i++ {
		formats[strings.ToLower(i.String())] = i
	}

	// add alternative format names.
	formats["rdf-xml"] = RDFXML
	formats["rdfxml"] = RDFXML
	formats["jsonld"] = JSONLD
	formats["freeformat"] = FreeFormat
}
