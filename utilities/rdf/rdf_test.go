package rdf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsClass(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not a class", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML"},
		{name: "RDF property", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#Property", want: true},
		{name: "RDFS class", t: "http://www.w3.org/2000/01/rdf-schema#Class", want: true},
		{name: "RDFS resource", t: "http://www.w3.org/2000/01/rdf-schema#Resource", want: true},
		{name: "OWL class", t: "http://www.w3.org/2002/07/owl#Class", want: true},
		{name: "OWL thing", t: "http://www.w3.org/2002/07/owl#Thing", want: true},
		{name: "SKOS concept", t: "http://www.w3.org/2004/02/skos/core#Concept", want: true},
		{name: "FTV attribute", t: "https://ftv.nl/rdf/ftv#Attribute", want: true},
		{name: "FTV entity-attribute", t: "https://ftv.nl/rdf/ftv#entityAttribute", want: true},
		{name: "FTV entity", t: "https://ftv.nl/rdf/ftv#Entity", want: true},
		{name: "FTV relation", t: "https://ftv.nl/rdf/ftv#Relation", want: true},
		{name: "FTV principal", t: "https://ftv.nl/rdf/ftv#principal", want: true},
		{name: "FTV resource", t: "https://ftv.nl/rdf/ftv#resource", want: true},
		{name: "FTV context", t: "https://ftv.nl/rdf/ftv#context", want: true},
		{name: "FTV subject", t: "https://ftv.nl/rdf/ftv#subject", want: true},
		{name: "FTV object", t: "https://ftv.nl/rdf/ftv#object", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsClass(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsLabel(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not a label", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML"},
		{name: "RDFS label", t: "http://www.w3.org/2000/01/rdf-schema#label", want: true},
		{name: "SKOS prefLabel", t: "http://www.w3.org/2004/02/skos/core#prefLabel", want: true},
		{name: "SKOS altLabel", t: "http://www.w3.org/2004/02/skos/core#altLabel", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsLabel(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsTitle(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not a title", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML"},
		{name: "DC terms title", t: "http://purl.org/dc/terms/title", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsTitle(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsList(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not a list", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML"},
		{name: "RDF bag", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#Bag", want: true},
		{name: "RDF seq", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#Seq", want: true},
		{name: "RDF list", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#List", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsList(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsAttribute(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not an attribute", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML"},
		{name: "FTV attribute", t: "https://ftv.nl/rdf/ftv#Attribute", want: true},
		{name: "FTV entity-attribute", t: "https://ftv.nl/rdf/ftv#entityAttribute", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsAttribute(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsEntity(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not an entity", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML"},
		{name: "FTV entity", t: "https://ftv.nl/rdf/ftv#Entity", want: true},
		{name: "FTV principal", t: "https://ftv.nl/rdf/ftv#principal", want: true},
		{name: "FTV resource", t: "https://ftv.nl/rdf/ftv#resource", want: true},
		{name: "FTV context", t: "https://ftv.nl/rdf/ftv#context", want: true},
		{name: "FTV subject", t: "https://ftv.nl/rdf/ftv#subject", want: true},
		{name: "FTV object", t: "https://ftv.nl/rdf/ftv#object", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsEntity(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsRelation(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not a relation", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML"},
		{name: "FTV relation", t: "https://ftv.nl/rdf/ftv#Relation", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsRelation(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsValue(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not a value", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#label"},
		{name: "RDF value", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#value", want: true},
		{name: "RDF property", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#Property", want: true},
		{name: "FTV attribute key", t: "https://ftv.nl/rdf/ftv#attributeKey", want: true},
		{name: "FTV entity type", t: "https://ftv.nl/rdf/ftv#entityType", want: true},
		{name: "FTV entity id", t: "https://ftv.nl/rdf/ftv#entityID", want: true},
		{name: "FTV action", t: "https://ftv.nl/rdf/ftv#action", want: true},
		{name: "FTV relation type", t: "https://ftv.nl/rdf/ftv#relationType", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsValue(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsLiteral(t *testing.T) {
	testCases := []struct {
		name string
		t    string
		want bool
	}{
		{name: "empty"},
		{name: "bad identifier", t: "abc"},
		{name: "not a literal", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#label"},
		{name: "FTV attribute key", t: "https://ftv.nl/rdf/ftv#attributeKey", want: true},
		{name: "FTV entity type", t: "https://ftv.nl/rdf/ftv#entityType", want: true},
		{name: "FTV entity id", t: "https://ftv.nl/rdf/ftv#entityID", want: true},
		{name: "FTV action", t: "https://ftv.nl/rdf/ftv#action", want: true},
		{name: "FTV relation type", t: "https://ftv.nl/rdf/ftv#relationType", want: true},
		{name: "RDF html", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML", want: true},
		{name: "RDF xml", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral", want: true},
		{name: "RDF json", t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#JSON", want: true},
		{name: "XSD any", t: "http://www.w3.org/2001/XMLSchema#anyType", want: true},
		{name: "XSD double", t: "http://www.w3.org/2001/XMLSchema#double", want: true},
		{name: "XSD long", t: "http://www.w3.org/2001/XMLSchema#long", want: true},
		{name: "XSD month", t: "http://www.w3.org/2001/XMLSchema#gMonth", want: true},
		{name: "XSD yearMonth", t: "http://www.w3.org/2001/XMLSchema#gYearMonth", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsLiteral(tc.t)
			assert.Equal(t, tc.want, got)
		})
	}
}
