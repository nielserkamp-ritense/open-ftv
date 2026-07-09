package models

import "strings"

// Language represents a policy language.
type Language uint8

// List of supported policy languages.
const (
	XACML Language = iota + 1
	ODRL
	REGO
	CEDAR
	CERBOS
	OPENFGA
	ODRLGEO

	langFirst = XACML
	langLast  = ODRLGEO
)

// String implements the Stringer interface.
func (l Language) String() string {
	switch l {
	case XACML:
		return "XACML"
	case ODRL:
		return "ODRL"
	case REGO:
		return "OPA/Rego"
	case CEDAR:
		return "Cedar"
	case CERBOS:
		return "Cerbos/CEL"
	case OPENFGA:
		return "OpenFGA"
	case ODRLGEO:
		return "ODRL-Geo"
	default:
		return "<unknown>"
	}
}

// Language returns the common language name for the policy language.
func (l Language) Language() string {
	switch l {
	case XACML:
		return "xacml"
	case ODRL:
		return "odrl"
	case REGO:
		return "rego"
	case CEDAR:
		return "cedar"
	case CERBOS:
		return "cerbos"
	case OPENFGA:
		return "openfga"
	case ODRLGEO:
		return "odrl-geo"
	default:
		return "<unknown>"
	}
}

// LanguageFromString returns the corresponding Language type from the given input.
func LanguageFromString(in string) Language {
	return languages[strings.ToLower(in)]
}

var languages = make(map[string]Language, langLast+8)

func init() {
	for i := langFirst; i <= langLast; i++ {
		languages[strings.ToLower(i.String())] = i
	}

	// add alternate language names.
	languages["opa"] = REGO
	languages["rego"] = REGO
	languages["opa-rego"] = REGO
	languages["oparego"] = REGO
	languages["cel"] = CERBOS
	languages["cerbos"] = CERBOS
	languages["cerbos-cel"] = CERBOS
	languages["cerboscel"] = CERBOS
	languages["open-fga"] = OPENFGA

	// ODRL-Geo-NL engine aliases. The engine consumes policies stored under the
	// "odrl" PAP language, so both PAP_LANGUAGE=odrl and =odrl-geo select it.
	languages["odrl-geo"] = ODRLGEO
	languages["odrlgeo"] = ODRLGEO
	languages["odrl-geo-nl"] = ODRLGEO
	languages["geo"] = ODRLGEO
}
