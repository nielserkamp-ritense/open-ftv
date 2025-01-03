package rdf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrefixUri(t *testing.T) {
	testCases := []struct {
		name   string
		prefix string
		want   string
	}{
		{name: "empty"},
		{name: "invalid", prefix: "abc"},
		{name: "DC terms", prefix: PrefixDCTerms, want: UriDCTerms},
		{name: "FTV", prefix: PrefixFTV, want: UriFTV},
		{name: "OWL", prefix: PrefixOWL, want: UriOWL},
		{name: "RDF", prefix: PrefixRDF, want: UriRDF},
		{name: "RDFS", prefix: PrefixRDFS, want: UriRDFS},
		{name: "SKOS", prefix: PrefixSKOS, want: UriSKOS},
		{name: "XSD", prefix: PrefixXSD, want: UriXSD},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := PrefixUri[tc.prefix]
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestUriPrefix(t *testing.T) {
	testCases := []struct {
		name   string
		prefix string
		want   string
	}{
		{name: "empty"},
		{name: "invalid", prefix: "abc"},
		{name: "DC terms", prefix: UriDCTerms, want: PrefixDCTerms},
		{name: "FTV", prefix: UriFTV, want: PrefixFTV},
		{name: "OWL", prefix: UriOWL, want: PrefixOWL},
		{name: "RDF", prefix: UriRDF, want: PrefixRDF},
		{name: "RDFS", prefix: UriRDFS, want: PrefixRDFS},
		{name: "SKOS", prefix: UriSKOS, want: PrefixSKOS},
		{name: "XSD", prefix: UriXSD, want: PrefixXSD},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := UriPrefix[tc.prefix]
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestConstants(t *testing.T) {
	testCases := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "RDFS comment",
			id:   "http://www.w3.org/2000/01/rdf-schema#comment",
			want: Comment,
		},
		{
			name: "RDFS domain",
			id:   "http://www.w3.org/2000/01/rdf-schema#domain",
			want: Domain,
		},
		{
			name: "RDFS literal",
			id:   "http://www.w3.org/2000/01/rdf-schema#Literal",
			want: Literal,
		},
		{
			name: "RDFS range",
			id:   "http://www.w3.org/2000/01/rdf-schema#range",
			want: Range,
		},
		{
			name: "RDFS sub class",
			id:   "http://www.w3.org/2000/01/rdf-schema#subClassOf",
			want: SubClass,
		},
		{
			name: "RDFS sub property",
			id:   "http://www.w3.org/2000/01/rdf-schema#subPropertyOf",
			want: SubProperty,
		},
		{
			name: "DC terms defined",
			id:   "http://purl.org/dc/terms/isDefinedBy",
			want: DCTermsDefined,
		},
		{
			name: "DC terms requires",
			id:   "http://purl.org/dc/terms/requires",
			want: DCTermsRequires,
		},
		{
			name: "DC terms required",
			id:   "http://purl.org/dc/terms/isRequiredBy",
			want: DCTermsRequired,
		},
		{
			name: "SKOS example",
			id:   "http://www.w3.org/2004/02/skos/core#example",
			want: SKOSExample,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.id)
		})
	}
}
