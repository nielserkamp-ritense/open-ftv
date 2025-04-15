package rdf

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/xsd"
)

func TestPrefixUri(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		prefix string
		want   string
	}{
		{name: "empty"},
		{name: "invalid", prefix: "abc"},
		{name: "DC terms", prefix: PrefixDCTerms, want: URIDCTerms},
		{name: "FTV", prefix: PrefixFTV, want: URIFTV},
		{name: "OWL", prefix: PrefixOWL, want: URIOWL},
		{name: "RDF", prefix: PrefixRDF, want: URIRDF},
		{name: "RDFS", prefix: PrefixRDFS, want: URIRDFS},
		{name: "SKOS", prefix: PrefixSKOS, want: URISKOS},
		{name: "XSD", prefix: xsd.Prefix, want: xsd.URI},
		{name: "VCard", prefix: PrefixVCard, want: URIVCard},
		{name: "Schema", prefix: PrefixSchema, want: URISchema},
		{name: "ODRL", prefix: PrefixODRL, want: URIODRL},
		{name: "CC", prefix: PrefixCC, want: URICC},
		{name: "FOAF", prefix: PrefixFOAF, want: URIFOAF},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := PrefixURI[tc.prefix]
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestUriPrefix(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		prefix string
		want   string
	}{
		{name: "empty"},
		{name: "invalid", prefix: "abc"},
		{name: "DC terms", prefix: URIDCTerms, want: PrefixDCTerms},
		{name: "FTV", prefix: URIFTV, want: PrefixFTV},
		{name: "OWL", prefix: URIOWL, want: PrefixOWL},
		{name: "RDF", prefix: URIRDF, want: PrefixRDF},
		{name: "RDFS", prefix: URIRDFS, want: PrefixRDFS},
		{name: "SKOS", prefix: URISKOS, want: PrefixSKOS},
		{name: "XSD", prefix: xsd.URI, want: xsd.Prefix},
		{name: "FOAF", prefix: URIFOAF, want: PrefixFOAF},
		{name: "CC", prefix: URICC, want: PrefixCC},
		{name: "ODRL", prefix: URIODRL, want: PrefixODRL},
		{name: "VCard", prefix: URIVCard, want: PrefixVCard},
		{name: "Schema", prefix: URISchema, want: PrefixSchema},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := URIPrefix[tc.prefix]
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestConstants(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		id   string
		want string
	}{
		{name: "RDFS comment", id: "http://www.w3.org/2000/01/rdf-schema#comment", want: Comment},
		{name: "RDFS domain", id: "http://www.w3.org/2000/01/rdf-schema#domain", want: Domain},
		{name: "RDFS literal", id: "http://www.w3.org/2000/01/rdf-schema#Literal", want: Literal},
		{name: "RDFS range", id: "http://www.w3.org/2000/01/rdf-schema#range", want: Range},
		{name: "RDFS sub class", id: "http://www.w3.org/2000/01/rdf-schema#subClassOf", want: SubClass},
		{name: "RDFS sub property", id: "http://www.w3.org/2000/01/rdf-schema#subPropertyOf", want: SubProperty},
		{name: "DC terms defined", id: "http://purl.org/dc/terms/isDefinedBy", want: DCTermsDefined},
		{name: "DC terms requires", id: "http://purl.org/dc/terms/requires", want: DCTermsRequires},
		{name: "DC terms required", id: "http://purl.org/dc/terms/isRequiredBy", want: DCTermsRequired},
		{name: "SKOS example", id: "http://www.w3.org/2004/02/skos/core#example", want: SKOSExample},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tc.id)
		})
	}
}
